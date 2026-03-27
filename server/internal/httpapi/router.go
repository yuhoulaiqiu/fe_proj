package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"community-help-hub-server/internal/domain"
	"community-help-hub-server/internal/store"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {
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

		userID, role, err := store.VerifyUser(db, req.Username, req.Password)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "invalid_credentials")
			return
		}

		token, expiresAt, err := store.CreateSession(db, userID)
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

		userID, err := store.CreateUser(db, req.Username, req.Password, "user")
		if err != nil {
			if store.IsDuplicateUsername(err) {
				writeError(c, http.StatusConflict, "username_taken")
				return
			}
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		token, expiresAt, err := store.CreateSession(db, userID)
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
	auth.Use(RequireAuth(db))
	{
		auth.GET("/me", func(c *gin.Context) {
			userID := c.MustGet("user_id").(int64)
			username, role, err := store.GetUserProfile(db, userID)
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
		_ = store.ReconcileActivities(db, time.Now())
		category := strings.TrimSpace(c.Query("category"))
		status := strings.TrimSpace(c.Query("status"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := store.ListActivities(db, category, status, keyword, page, pageSize)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[domain.Activity]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})

	r.GET("/api/activities/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		_, _ = store.ReconcileActivityStatusByID(db, id, time.Now())
		it, err := store.GetActivity(db, id)
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
	user.Use(RequireAuth(db))
	{
		user.GET("/notifications", func(c *gin.Context) {
			page, pageSize := parsePage(c)
			userID := c.MustGet("user_id").(int64)
			items, total, err := store.ListUserNotifications(db, userID, page, pageSize)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, listResponse[domain.Notification]{Items: items, Total: total, Page: page, PageSize: pageSize})
		})

		user.POST("/notifications/:id/read", func(c *gin.Context) {
			notificationID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			if err := store.MarkNotificationRead(db, userID, notificationID); err != nil {
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
			var it domain.Activity
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
			if err := store.UpdateActivityByOwner(db, activityID, userID, it); err != nil {
				switch {
				case errors.Is(err, store.ErrActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, store.ErrForbidden):
					writeError(c, http.StatusForbidden, "forbidden")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}
			_, _ = store.ReconcileActivityStatusByID(db, activityID, time.Now())
			updated, err := store.GetActivity(db, activityID)
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
			if err := store.DeleteActivityByOwner(db, activityID, userID); err != nil {
				switch {
				case errors.Is(err, store.ErrActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, store.ErrForbidden):
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
			items, err := store.ListActivityRegistrationsByOwner(db, activityID, userID)
			if err != nil {
				switch {
				case errors.Is(err, store.ErrActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, store.ErrForbidden):
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
			items, err := store.ListActivityRegistrationsByOwner(db, activityID, userID)
			if err != nil {
				switch {
				case errors.Is(err, store.ErrActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, store.ErrForbidden):
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
			if err := store.RegisterActivity(db, activityID, userID); err != nil {
				switch {
				case errors.Is(err, store.ErrActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, store.ErrAlreadyRegistered):
					writeError(c, http.StatusConflict, "already_registered")
					return
				case errors.Is(err, store.ErrRegistrationClosed):
					writeError(c, http.StatusConflict, "registration_closed")
					return
				case errors.Is(err, store.ErrActivityTimeInvalid):
					writeError(c, http.StatusBadRequest, "activity_time_invalid")
					return
				default:
					writeError(c, http.StatusInternalServerError, "server_error")
					return
				}
			}
			c.JSON(http.StatusOK, map[string]any{"ok": true})
		})

		user.DELETE("/activities/:id/register", func(c *gin.Context) {
			activityID, ok := parseID(c.Param("id"))
			if !ok {
				writeError(c, http.StatusBadRequest, "invalid_id")
				return
			}
			userID := c.MustGet("user_id").(int64)
			if err := store.CancelActivityRegistration(db, activityID, userID); err != nil {
				switch {
				case errors.Is(err, store.ErrActivityNotFound):
					writeError(c, http.StatusNotFound, "not_found")
					return
				case errors.Is(err, store.ErrNotRegistered):
					writeError(c, http.StatusNotFound, "not_registered")
					return
				case errors.Is(err, store.ErrCancellationClosed):
					writeError(c, http.StatusConflict, "cancellation_closed")
					return
				case errors.Is(err, store.ErrActivityTimeInvalid):
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
			items, err := store.ListUserRegisteredActivities(db, userID)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, items)
		})

		user.GET("/user/activities/published", func(c *gin.Context) {
			userID := c.MustGet("user_id").(int64)
			items, err := store.ListUserPublishedActivities(db, userID)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			c.JSON(http.StatusOK, items)
		})

		user.POST("/activities", func(c *gin.Context) {
			var it domain.Activity
			if err := c.ShouldBindJSON(&it); err != nil {
				writeError(c, http.StatusBadRequest, "invalid_json")
				return
			}
			it.UserID = c.MustGet("user_id").(int64)
			id, err := store.CreateActivity(db, it)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "server_error")
				return
			}
			created, err := store.GetActivity(db, id)
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
		items, total, err := store.ListServices(db, category, keyword, page, pageSize)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[domain.Service]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})

	r.GET("/api/services/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := store.GetService(db, id)
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
		items, total, err := store.ListLostItems(db, itemType, status, keyword, page, pageSize, false)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[domain.LostItem]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})

	r.GET("/api/lost-items/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := store.GetLostItem(db, id)
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
	admin.Use(RequireAdmin(db))
	admin.GET("/services", func(c *gin.Context) {
		category := strings.TrimSpace(c.Query("category"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := store.ListServices(db, category, keyword, page, pageSize)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[domain.Service]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})
	admin.POST("/services", func(c *gin.Context) {
		var it domain.Service
		if err := c.ShouldBindJSON(&it); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		it.Name = strings.TrimSpace(it.Name)
		if it.Name == "" {
			writeError(c, http.StatusBadRequest, "missing_name")
			return
		}
		id, err := store.CreateService(db, it)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		created, err := store.GetService(db, id)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusCreated, created)
	})
	admin.GET("/services/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := store.GetService(db, id)
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
	admin.PUT("/services/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		var it domain.Service
		if err := c.ShouldBindJSON(&it); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		it.Name = strings.TrimSpace(it.Name)
		if it.Name == "" {
			writeError(c, http.StatusBadRequest, "missing_name")
			return
		}
		if err := store.UpdateService(db, id, it); err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		updated, err := store.GetService(db, id)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, updated)
	})
	admin.DELETE("/services/:id", func(c *gin.Context) {
		id, ok := parseID(c.Param("id"))
		if !ok {
			writeError(c, http.StatusBadRequest, "invalid_id")
			return
		}
		if err := store.DeleteService(db, id); err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
	admin.GET("/lost-items", func(c *gin.Context) {
		itemType := strings.TrimSpace(c.Query("type"))
		status := strings.TrimSpace(c.Query("status"))
		keyword := strings.TrimSpace(c.Query("keyword"))
		page, pageSize := parsePage(c)
		items, total, err := store.ListLostItems(db, itemType, status, keyword, page, pageSize, false)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, listResponse[domain.LostItem]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})
	admin.POST("/lost-items", func(c *gin.Context) {
		var it domain.LostItem
		if err := c.ShouldBindJSON(&it); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		id, err := store.CreateLostItem(db, it)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		created, err := store.GetLostItem(db, id)
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
		it, err := store.GetLostItem(db, id)
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
		var it domain.LostItem
		if err := c.ShouldBindJSON(&it); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_json")
			return
		}
		if err := store.UpdateLostItem(db, id, it); err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		updated, err := store.GetLostItem(db, id)
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
		if err := store.DeleteLostItem(db, id); err != nil {
			writeError(c, http.StatusInternalServerError, "server_error")
			return
		}
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
}

func WithCORS(allowed []string) gin.HandlerFunc {
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

func RequireAuth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		userID, err := store.ValidateSession(db, token)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func RequireAdmin(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		userID, err := store.ValidateSession(db, token)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		_, role, err := store.GetUserProfile(db, userID)
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

func bearerToken(authHeader string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	tok := strings.TrimSpace(parts[1])
	if tok == "" {
		return "", false
	}
	return tok, true
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

type listResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
	User      struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

type meResponse struct {
	User struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}
