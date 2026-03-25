package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func registerRoutes(r *gin.Engine, db *sql.DB) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"ok": true, "time": time.Now().Format(time.RFC3339)})
	})

	r.POST("/api/auth/login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			writeError(c, http.StatusBadRequest, "missing_credentials")
			return
		}

		userID, role, err := verifyUser(db, req.Username, req.Password)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "invalid_credentials")
			return
		}

		token, expiresAt, err := createSession(db, userID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}

		var res loginResponse
		res.Token = token
		res.ExpiresAt = expiresAt.Format(time.RFC3339)
		res.User.ID = userID
		res.User.Username = req.Username
		res.User.Role = role
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/auth/register", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			writeError(c, http.StatusBadRequest, "missing_credentials")
			return
		}
		if len(req.Password) < 6 {
			writeError(c, http.StatusBadRequest, "weak_password")
			return
		}

		userID, err := createUser(db, req.Username, req.Password, "user")
		if err != nil {
			if isDuplicateUsername(err) {
				writeError(c, http.StatusConflict, "username_taken")
				return
			}
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		token, expiresAt, err := createSession(db, userID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		var res loginResponse
		res.Token = token
		res.ExpiresAt = expiresAt.Format(time.RFC3339)
		res.User.ID = userID
		res.User.Username = req.Username
		res.User.Role = "user"
		c.JSON(http.StatusCreated, res)
	})

	auth := r.Group("/api/auth")
	auth.Use(requireAuth(db))
	{
		auth.GET("/me", func(c *gin.Context) {
			userID := c.MustGet("user_id").(int64)
			username, role, err := getUserProfile(db, userID)
			if errors.Is(err, sql.ErrNoRows) {
				writeError(c, http.StatusUnauthorized, "unauthorized")
				return
			}
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			var res meResponse
			res.User.ID = userID
			res.User.Username = username
			res.User.Role = role
			c.JSON(http.StatusOK, res)
		})
	}

	r.GET("/api/activities", func(c *gin.Context) {
		_ = reconcileActivities(db, time.Now())
		category := strings.TrimSpace(c.Query("category"))
		status := strings.TrimSpace(c.Query("status"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := listActivities(db, category, status, keyword, page, pageSize)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[Activity]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})

	r.GET("/api/activities/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		_, _ = reconcileActivityStatusByID(db, id, time.Now())
		it, err := getActivity(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(c, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, it)
	})

	user := r.Group("/api")
	user.Use(requireAuth(db))
	{
		user.GET("/notifications", func(c *gin.Context) {
			page, pageSize := parsePage(c)
			userID := c.MustGet("user_id").(int64)
			items, total, err := listUserNotifications(db, userID, page, pageSize)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, listResponse[Notification]{Items: items, Total: total, Page: page, PageSize: pageSize})
		})

		user.POST("/notifications/:id/read", func(c *gin.Context) {
			notificationID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			if err := markNotificationRead(db, userID, notificationID); err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, map[string]any{"ok": true})
		})

		user.PUT("/activities/:id", func(c *gin.Context) {
			activityID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			var it Activity
			if err := c.ShouldBindJSON(&it); err != nil {
				writeError(c, http.StatusBadRequest, "invalid_json")
				return
			}
			it.Title = strings.TrimSpace(it.Title)
			it.Category = strings.TrimSpace(it.Category)
			it.Status = strings.TrimSpace(it.Status)
			if it.Title == "" {
				writeError(c, http.StatusBadRequest, "missing_title")
				return
			}
			userID := c.MustGet("user_id").(int64)
			if err := updateActivityByOwner(db, activityID, userID, it); err != nil {
				switch {
				case errors.Is(err, errActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, errForbidden):
					writeError(c, http.StatusForbidden, "forbidden")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}
			_, _ = reconcileActivityStatusByID(db, activityID, time.Now())
			updated, err := getActivity(db, activityID)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, updated)
		})

		user.DELETE("/activities/:id", func(c *gin.Context) {
			activityID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			if err := deleteActivityByOwner(db, activityID, userID); err != nil {
				switch {
				case errors.Is(err, errActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, errForbidden):
					writeError(c, http.StatusForbidden, "forbidden")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}
			c.JSON(http.StatusOK, map[string]any{"ok": true})
		})

		user.GET("/activities/:id/registrations", func(c *gin.Context) {
			activityID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			items, err := listActivityRegistrationsByOwner(db, activityID, userID)
			if err != nil {
				switch {
				case errors.Is(err, errActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, errForbidden):
					writeError(c, http.StatusForbidden, "forbidden")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}
			c.JSON(http.StatusOK, items)
		})

		user.GET("/activities/:id/registrations.csv", func(c *gin.Context) {
			activityID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			items, err := listActivityRegistrationsByOwner(db, activityID, userID)
			if err != nil {
				switch {
				case errors.Is(err, errActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, errForbidden):
					writeError(c, http.StatusForbidden, "forbidden")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}

			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			_ = w.Write([]string{"registration_id", "activity_id", "user_id", "username", "status", "created_at"})
			for _, it := range items {
				_ = w.Write([]string{
					strconv.FormatInt(it.ID, 10),
					strconv.FormatInt(it.ActivityID, 10),
					strconv.FormatInt(it.UserID, 10),
					it.Username,
					it.Status,
					it.CreatedAt,
				})
			}
			w.Flush()
			if err := w.Error(); err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}

			c.Header("Content-Type", "text/csv; charset=utf-8")
			c.Header("Content-Disposition", "attachment; filename=\"activity_registrations.csv\"")
			c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
		})

		user.POST("/activities/:id/register", func(c *gin.Context) {
			activityID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			if err := registerActivity(db, activityID, userID); err != nil {
				switch {
				case errors.Is(err, errActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, errAlreadyRegistered):
					writeError(c, http.StatusConflict, "already_registered")
					return
				case errors.Is(err, errRegistrationClosed):
					writeError(c, http.StatusConflict, "registration_closed")
					return
				case errors.Is(err, errActivityTimeInvalid):
					writeError(c, http.StatusBadRequest, "activity_time_invalid")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}
			c.JSON(http.StatusOK, map[string]any{"ok": true})
		})

		user.GET("/user/activities/registered", func(c *gin.Context) {
			userID := c.MustGet("user_id").(int64)
			items, err := listUserRegisteredActivities(db, userID)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, items)
		})

		user.GET("/user/activities/published", func(c *gin.Context) {
			userID := c.MustGet("user_id").(int64)
			items, err := listUserPublishedActivities(db, userID)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, items)
		})

		user.POST("/activities", func(c *gin.Context) {
			var it Activity
			if err := c.ShouldBindJSON(&it); err != nil {
				writeError(c, http.StatusBadRequest, "invalid_json")
				return
			}
			it.UserID = c.MustGet("user_id").(int64)
			id, err := createActivity(db, it)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			created, err := getActivity(db, id)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusCreated, created)
		})
	}

	r.GET("/api/services", func(c *gin.Context) {
		category := strings.TrimSpace(c.Query("category"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := listServices(db, category, keyword, page, pageSize)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[Service]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})

	r.GET("/api/services/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := getService(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(c, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, it)
	})

	r.GET("/api/lost-items", func(c *gin.Context) {
		itemType := strings.TrimSpace(c.Query("type"))
		status := strings.TrimSpace(c.Query("status"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := listLostItems(db, itemType, status, keyword, page, pageSize, false)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[LostItem]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})

	r.GET("/api/lost-items/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := getLostItem(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(c, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, it)
	})

	admin := r.Group("/api/admin")
	admin.Use(requireAdmin(db))
	admin.GET("/lost-items", func(c *gin.Context) {
		itemType := strings.TrimSpace(c.Query("type"))
		status := strings.TrimSpace(c.Query("status"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := listLostItems(db, itemType, status, keyword, page, pageSize, false)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[LostItem]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})
	admin.POST("/lost-items", func(c *gin.Context) {
		var it LostItem
		if err := c.ShouldBindJSON(&it); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		id, err := createLostItem(db, it)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		created, err := getLostItem(db, id)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusCreated, created)
	})
	admin.GET("/lost-items/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := getLostItem(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(c, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, it)
	})
	admin.PUT("/lost-items/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		var it LostItem
		if err := c.ShouldBindJSON(&it); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		if err := updateLostItem(db, id, it); err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		updated, err := getLostItem(db, id)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, updated)
	})
	admin.DELETE("/lost-items/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		if err := deleteLostItem(db, id); err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
}

func withCORS(allowed []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := ""
		if len(allowed) == 0 {
			if origin != "" {
				allowOrigin = origin
			}
		} else if origin != "" {
			for _, a := range allowed {
				if a == "*" || strings.EqualFold(a, origin) {
					allowOrigin = a
					break
				}
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func requireAuth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		userID, err := validateSession(db, token)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func requireAdmin(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		userID, err := validateSession(db, token)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		_, role, err := getUserProfile(db, userID)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		if role != "admin" {
			writeError(c, http.StatusForbidden, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

func writeError(c *gin.Context, status int, code string) {
	c.JSON(status, map[string]any{"message": code})
}

func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func parseID(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}
