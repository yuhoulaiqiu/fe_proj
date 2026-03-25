package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func newTestRouter(t *testing.T, db *sql.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withCORS(nil))
	registerRoutes(r, db)
	return r
}

func TestHealth(t *testing.T) {
	r := newTestRouter(t, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, password, password_hash, role FROM users WHERE username = \\?").
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password", "password_hash", "role"}).AddRow(int64(1), "admin123", "", "admin"))

	mock.ExpectExec("INSERT INTO sessions \\(token, user_id, expires_at, created_at\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := newTestRouter(t, db)

	body, _ := json.Marshal(map[string]any{"username": "admin", "password": "admin123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var res loginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Token == "" || res.User.ID != 1 || res.User.Role != "admin" {
		t.Fatalf("unexpected response: %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLoginLegacyHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT id, password, password_hash, role FROM users WHERE username = \\?").
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password", "password_hash", "role"}).AddRow(int64(1), "", string(hash), "admin"))

	mock.ExpectExec("INSERT INTO sessions \\(token, user_id, expires_at, created_at\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := newTestRouter(t, db)

	body, _ := json.Marshal(map[string]any{"username": "admin", "password": "admin123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var res loginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Token == "" || res.User.ID != 1 || res.User.Role != "admin" {
		t.Fatalf("unexpected response: %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListActivitiesEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM activities").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(1)"}).AddRow(int64(0)))

	mock.ExpectQuery("SELECT id, title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at FROM activities").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "category", "status", "user_id", "cover_url", "summary", "content", "location", "start_time", "end_time", "created_at"}))

	r := newTestRouter(t, db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/activities?page=1&pageSize=20", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var res struct {
		Items    []Activity `json:"items"`
		Total    int64      `json:"total"`
		Page     int        `json:"page"`
		PageSize int        `json:"pageSize"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Total != 0 || len(res.Items) != 0 {
		t.Fatalf("unexpected response: %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdminCreateLostItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	token := "t"
	expiresAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	mock.ExpectQuery("SELECT user_id, expires_at FROM sessions WHERE token = \\?").
		WithArgs(token).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow(int64(1), expiresAt))

	mock.ExpectQuery("SELECT username, role FROM users WHERE id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"username", "role"}).AddRow("admin", "admin"))

	mock.ExpectExec("INSERT INTO lost_items").
		WithArgs("标题", "lost", "open", "地点", "时间", "描述", "联系方式", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(5, 1))

	mock.ExpectQuery("SELECT id, title, item_type, status, location, occurred_at, description, contact, created_at, updated_at FROM lost_items WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "item_type", "status", "location", "occurred_at", "description", "contact", "created_at", "updated_at"}).
			AddRow(int64(5), "标题", "lost", "open", "地点", "时间", "描述", "联系方式", "c", "u"))

	r := newTestRouter(t, db)

	body, _ := json.Marshal(map[string]any{
		"title":       "标题",
		"itemType":    "lost",
		"status":      "open",
		"location":    "地点",
		"occurredAt":  "时间",
		"description": "描述",
		"contact":     "联系方式",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/lost-items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
