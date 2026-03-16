package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type ctxKey string

const ctxUserIDKey ctxKey = "user_id"

type Activity struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	CoverURL  string `json:"coverUrl,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Content   string `json:"content,omitempty"`
	Location  string `json:"location,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type Service struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type LostItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	ItemType    string `json:"itemType,omitempty"`
	Status      string `json:"status,omitempty"`
	Location    string `json:"location,omitempty"`
	OccurredAt  string `json:"occurredAt,omitempty"`
	Description string `json:"description,omitempty"`
	Contact     string `json:"contact,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
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

func main() {
	port := envOr("PORT", "8080")
	dbPath := envOr("DB_PATH", "./data/app.db")
	origins := parseOrigins(os.Getenv("CORS_ORIGINS"))

	if err := os.MkdirAll("./data", 0o755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}
	if err := ensureAdminUser(db); err != nil {
		log.Fatal(err)
	}
	if err := seedIfEmpty(db); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "time": time.Now().Format(time.RFC3339)})
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		var req loginRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "missing_credentials")
			return
		}

		userID, role, err := verifyUser(db, req.Username, req.Password)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}

		token, expiresAt, err := createSession(db, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}

		var res loginResponse
		res.Token = token
		res.ExpiresAt = expiresAt.Format(time.RFC3339)
		res.User.ID = userID
		res.User.Username = req.Username
		res.User.Role = role
		writeJSON(w, http.StatusOK, res)
	})

	mux.HandleFunc("/api/activities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
		page, pageSize := parsePage(r)
		items, total, err := listActivities(db, keyword, page, pageSize)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}
		writeJSON(w, http.StatusOK, listResponse[Activity]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})
	mux.HandleFunc("/api/activities/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		id, ok := parseID(strings.TrimPrefix(r.URL.Path, "/api/activities/"))
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := getActivity(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}
		writeJSON(w, http.StatusOK, it)
	})

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		category := strings.TrimSpace(r.URL.Query().Get("category"))
		keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
		page, pageSize := parsePage(r)
		items, total, err := listServices(db, category, keyword, page, pageSize)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}
		writeJSON(w, http.StatusOK, listResponse[Service]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})
	mux.HandleFunc("/api/services/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		id, ok := parseID(strings.TrimPrefix(r.URL.Path, "/api/services/"))
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := getService(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}
		writeJSON(w, http.StatusOK, it)
	})

	mux.HandleFunc("/api/lost-items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		itemType := strings.TrimSpace(r.URL.Query().Get("type"))
		status := strings.TrimSpace(r.URL.Query().Get("status"))
		keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
		page, pageSize := parsePage(r)
		items, total, err := listLostItems(db, itemType, status, keyword, page, pageSize, false)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}
		writeJSON(w, http.StatusOK, listResponse[LostItem]{Items: items, Total: total, Page: page, PageSize: pageSize})
	})
	mux.HandleFunc("/api/lost-items/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		id, ok := parseID(strings.TrimPrefix(r.URL.Path, "/api/lost-items/"))
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		it, err := getLostItem(db, id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return
		}
		writeJSON(w, http.StatusOK, it)
	})

	admin := requireAdmin(db, muxAdmin(db))
	mux.Handle("/api/admin/", http.StripPrefix("/api/admin", admin))

	handler := withCORS(origins, mux)
	addr := ":" + port
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func muxAdmin(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/lost-items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			itemType := strings.TrimSpace(r.URL.Query().Get("type"))
			status := strings.TrimSpace(r.URL.Query().Get("status"))
			keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
			page, pageSize := parsePage(r)
			items, total, err := listLostItems(db, itemType, status, keyword, page, pageSize, false)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			writeJSON(w, http.StatusOK, listResponse[LostItem]{Items: items, Total: total, Page: page, PageSize: pageSize})
		case http.MethodPost:
			var it LostItem
			if err := readJSON(r, &it); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_json")
				return
			}
			id, err := createLostItem(db, it)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			created, err := getLostItem(db, id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			writeJSON(w, http.StatusCreated, created)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	})

	mux.HandleFunc("/lost-items/", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(strings.TrimPrefix(r.URL.Path, "/lost-items/"))
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		switch r.Method {
		case http.MethodGet:
			it, err := getLostItem(db, id)
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "not_found")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			writeJSON(w, http.StatusOK, it)
		case http.MethodPut:
			var it LostItem
			if err := readJSON(r, &it); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_json")
				return
			}
			if err := updateLostItem(db, id, it); err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			updated, err := getLostItem(db, id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			writeJSON(w, http.StatusOK, updated)
		case http.MethodDelete:
			if err := deleteLostItem(db, id); err != nil {
				writeError(w, http.StatusInternalServerError, "server_error")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	})

	return mux
}

func withCORS(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
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
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseOrigins(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func envOr(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"message": code})
}

func parsePage(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
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

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			cover_url TEXT NOT NULL DEFAULT '',
			summary TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			start_time TEXT NOT NULL DEFAULT '',
			end_time TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS lost_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			item_type TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			occurred_at TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			contact TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			deleted_at TEXT
		);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func ensureAdminUser(db *sql.DB) error {
	adminUser := envOr("ADMIN_USERNAME", "admin")
	adminPass := envOr("ADMIN_PASSWORD", "admin123")
	var id int64
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", adminUser).Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec(
		"INSERT INTO users (username, password_hash, role, created_at) VALUES (?,?,?,?)",
		adminUser,
		string(hash),
		"admin",
		now,
	)
	return err
}

func verifyUser(db *sql.DB, username, password string) (int64, string, error) {
	var id int64
	var hash, role string
	err := db.QueryRow("SELECT id, password_hash, role FROM users WHERE username = ?", username).Scan(&id, &hash, &role)
	if err != nil {
		return 0, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, "", err
	}
	return id, role, nil
}

func createSession(db *sql.DB, userID int64) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(24 * time.Hour)
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at, created_at) VALUES (?,?,?,?)",
		token,
		userID,
		expiresAt.Format(time.RFC3339),
		now,
	)
	return token, expiresAt, err
}

func requireAdmin(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		token := strings.TrimSpace(parts[1])
		userID, err := validateSession(db, token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateSession(db *sql.DB, token string) (int64, error) {
	var userID int64
	var expiresAtStr string
	err := db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token = ?", token).Scan(&userID, &expiresAtStr)
	if err != nil {
		return 0, err
	}
	t, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return 0, err
	}
	if time.Now().After(t) {
		return 0, fmt.Errorf("expired")
	}
	return userID, nil
}

func listActivities(db *sql.DB, keyword string, page, pageSize int) ([]Activity, int64, error) {
	where := ""
	args := []any{}
	if keyword != "" {
		where = "WHERE title LIKE ? OR summary LIKE ?"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	var total int64
	qCount := "SELECT COUNT(1) FROM activities " + where
	if err := db.QueryRow(qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, title, cover_url, summary, content, location, start_time, end_time, created_at FROM activities " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Activity{}
	for rows.Next() {
		var it Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getActivity(db *sql.DB, id int64) (Activity, error) {
	var it Activity
	err := db.QueryRow(
		"SELECT id, title, cover_url, summary, content, location, start_time, end_time, created_at FROM activities WHERE id = ?",
		id,
	).Scan(&it.ID, &it.Title, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt)
	return it, err
}

func listServices(db *sql.DB, category, keyword string, page, pageSize int) ([]Service, int64, error) {
	clauses := []string{}
	args := []any{}
	if category != "" {
		clauses = append(clauses, "category = ?")
		args = append(args, category)
	}
	if keyword != "" {
		clauses = append(clauses, "(name LIKE ? OR description LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := db.QueryRow("SELECT COUNT(1) FROM services "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, name, category, phone, address, description, updated_at FROM services " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Service{}
	for rows.Next() {
		var it Service
		if err := rows.Scan(&it.ID, &it.Name, &it.Category, &it.Phone, &it.Address, &it.Description, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getService(db *sql.DB, id int64) (Service, error) {
	var it Service
	err := db.QueryRow(
		"SELECT id, name, category, phone, address, description, updated_at FROM services WHERE id = ?",
		id,
	).Scan(&it.ID, &it.Name, &it.Category, &it.Phone, &it.Address, &it.Description, &it.UpdatedAt)
	return it, err
}

func listLostItems(db *sql.DB, itemType, status, keyword string, page, pageSize int, includeDeleted bool) ([]LostItem, int64, error) {
	clauses := []string{}
	args := []any{}
	if !includeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	if itemType != "" {
		clauses = append(clauses, "item_type = ?")
		args = append(args, itemType)
	}
	if status != "" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if keyword != "" {
		clauses = append(clauses, "(title LIKE ? OR description LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := db.QueryRow("SELECT COUNT(1) FROM lost_items "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, title, item_type, status, location, occurred_at, description, contact, created_at, updated_at FROM lost_items " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []LostItem{}
	for rows.Next() {
		var it LostItem
		if err := rows.Scan(&it.ID, &it.Title, &it.ItemType, &it.Status, &it.Location, &it.OccurredAt, &it.Description, &it.Contact, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getLostItem(db *sql.DB, id int64) (LostItem, error) {
	var it LostItem
	err := db.QueryRow(
		"SELECT id, title, item_type, status, location, occurred_at, description, contact, created_at, updated_at FROM lost_items WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&it.ID, &it.Title, &it.ItemType, &it.Status, &it.Location, &it.OccurredAt, &it.Description, &it.Contact, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}

func createLostItem(db *sql.DB, it LostItem) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	if it.Status == "" {
		it.Status = "open"
	}
	res, err := db.Exec(
		`INSERT INTO lost_items (title, item_type, status, location, occurred_at, description, contact, created_at, updated_at) 
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(it.Title),
		strings.TrimSpace(it.ItemType),
		strings.TrimSpace(it.Status),
		strings.TrimSpace(it.Location),
		strings.TrimSpace(it.OccurredAt),
		strings.TrimSpace(it.Description),
		strings.TrimSpace(it.Contact),
		now,
		now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateLostItem(db *sql.DB, id int64, it LostItem) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec(
		`UPDATE lost_items 
		 SET title = ?, item_type = ?, status = ?, location = ?, occurred_at = ?, description = ?, contact = ?, updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		strings.TrimSpace(it.Title),
		strings.TrimSpace(it.ItemType),
		strings.TrimSpace(it.Status),
		strings.TrimSpace(it.Location),
		strings.TrimSpace(it.OccurredAt),
		strings.TrimSpace(it.Description),
		strings.TrimSpace(it.Contact),
		now,
		id,
	)
	return err
}

func deleteLostItem(db *sql.DB, id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec("UPDATE lost_items SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", now, now, id)
	return err
}

func seedIfEmpty(db *sql.DB) error {
	now := time.Now().Format(time.RFC3339)

	var c int64
	if err := db.QueryRow("SELECT COUNT(1) FROM activities").Scan(&c); err != nil {
		return err
	}
	if c == 0 {
		_, err := db.Exec(
			`INSERT INTO activities (title, cover_url, summary, content, location, start_time, end_time, created_at) VALUES
			(?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?)`,
			"周末垃圾分类科普站",
			"",
			"面向居民的垃圾分类小课堂与现场答疑。",
			"活动包含分类知识讲解、互动问答与小礼品发放，欢迎携家人参与。",
			"社区活动室",
			"2026-03-23 09:30",
			"2026-03-23 11:30",
			now,
			"义诊进社区",
			"",
			"联合社区医院开展基础健康检查。",
			"提供血压血糖测量、常见慢病咨询与健康宣教。",
			"社区广场",
			"2026-03-30 08:30",
			"2026-03-30 12:00",
			now,
			"敬老院探访志愿活动",
			"",
			"组织志愿者探访敬老院，陪伴聊天与节目表演。",
			"请提前报名并遵守探访规范，欢迎有才艺的居民参与。",
			"XX敬老院",
			"2026-04-06 14:00",
			"2026-04-06 17:00",
			now,
		)
		if err != nil {
			return err
		}
	}

	if err := db.QueryRow("SELECT COUNT(1) FROM services").Scan(&c); err != nil {
		return err
	}
	if c == 0 {
		_, err := db.Exec(
			`INSERT INTO services (name, category, phone, address, description, updated_at) VALUES
			(?,?,?,?,?,?),
			(?,?,?,?,?,?),
			(?,?,?,?,?,?),
			(?,?,?,?,?,?)`,
			"社区水电维修",
			"repair",
			"400-000-0001",
			"A区物业服务中心",
			"提供水电故障排查与上门维修预约。",
			now,
			"开锁与门窗维护",
			"repair",
			"400-000-0002",
			"B区沿街商铺 12 号",
			"身份证明核验后提供开锁服务，支持门窗五金更换。",
			now,
			"家政保洁预约",
			"housekeeping",
			"400-000-0003",
			"线上预约",
			"日常保洁、深度清洁与收纳整理服务。",
			now,
			"医保与社保办事指南",
			"guide",
			"12345",
			"社区服务大厅",
			"提供医保报销、社保查询、材料清单与办理流程。",
			now,
		)
		if err != nil {
			return err
		}
	}

	if err := db.QueryRow("SELECT COUNT(1) FROM lost_items").Scan(&c); err != nil {
		return err
	}
	if c == 0 {
		_, err := db.Exec(
			`INSERT INTO lost_items (title, item_type, status, location, occurred_at, description, contact, created_at, updated_at) VALUES
			(?,?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?,?)`,
			"地铁口捡到一串钥匙",
			"found",
			"open",
			"2号线 A 出口",
			"2026-03-15 18:40",
			"黑色钥匙扣，上面有小铃铛。",
			"张先生 138****0000",
			now,
			now,
			"丢失蓝色雨伞",
			"lost",
			"open",
			"社区活动室",
			"2026-03-16 10:20",
			"折叠伞，伞柄处有磨损。",
			"李女士 139****0000",
			now,
			now,
			"捡到学生证一张",
			"found",
			"claimed",
			"快递柜附近",
			"2026-03-14 20:10",
			"证件姓名：王同学（请核对信息领取）。",
			"物业前台 400-000-0000",
			now,
			now,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
