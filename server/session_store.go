package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

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
