package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"community-help-hub-server/internal/config"

	mysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func EnsureAdminUser(db *sql.DB) error {
	adminUser := config.EnvOr("ADMIN_USERNAME", "admin")
	adminPass := config.EnvOr("ADMIN_PASSWORD", "admin123")
	var id int64
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", adminUser).Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec(
		"INSERT INTO users (username, password, password_hash, role, created_at) VALUES (?,?,?,?,?)",
		adminUser,
		adminPass,
		"",
		"admin",
		now,
	)
	return err
}

func VerifyUser(db *sql.DB, username, password string) (int64, string, error) {
	var id int64
	var storedPass, hash, role string
	err := db.QueryRow("SELECT id, password, password_hash, role FROM users WHERE username = ?", username).Scan(&id, &storedPass, &hash, &role)
	if err != nil {
		return 0, "", err
	}
	if storedPass != "" {
		if storedPass != password {
			return 0, "", errors.New("invalid_credentials")
		}
		return id, role, nil
	}
	if hash == "" {
		return 0, "", errors.New("invalid_credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, "", errors.New("invalid_credentials")
	}
	return id, role, nil
}

func CreateUser(db *sql.DB, username, password, role string) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	res, err := db.Exec(
		"INSERT INTO users (username, password, password_hash, role, created_at) VALUES (?,?,?,?,?)",
		username,
		password,
		"",
		role,
		now,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func IsDuplicateUsername(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}

func GetUserProfile(db *sql.DB, userID int64) (string, string, error) {
	var username, role string
	err := db.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
	return username, role, err
}

func CreateSession(db *sql.DB, userID int64) (string, time.Time, error) {
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

func ValidateSession(db *sql.DB, token string) (int64, error) {
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
