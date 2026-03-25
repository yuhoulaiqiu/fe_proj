package main

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT NOT NULL AUTO_INCREMENT,
			username VARCHAR(191) NOT NULL,
			password VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL DEFAULT '',
			role VARCHAR(32) NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uk_users_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id BIGINT NOT NULL AUTO_INCREMENT,
			token VARCHAR(255) NOT NULL,
			user_id BIGINT NOT NULL,
			expires_at VARCHAR(64) NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uk_sessions_token (token),
			KEY idx_sessions_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS activities (
			id BIGINT NOT NULL AUTO_INCREMENT,
			title VARCHAR(255) NOT NULL,
			cover_url TEXT NOT NULL,
			summary TEXT NOT NULL,
			content MEDIUMTEXT NOT NULL,
			location VARCHAR(255) NOT NULL,
			start_time VARCHAR(64) NOT NULL,
			end_time VARCHAR(64) NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS services (
			id BIGINT NOT NULL AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			category VARCHAR(64) NOT NULL,
			phone VARCHAR(64) NOT NULL,
			address VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			updated_at VARCHAR(64) NOT NULL,
			PRIMARY KEY (id),
			KEY idx_services_category (category)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS lost_items (
			id BIGINT NOT NULL AUTO_INCREMENT,
			title VARCHAR(255) NOT NULL,
			item_type VARCHAR(32) NOT NULL,
			status VARCHAR(32) NOT NULL,
			location VARCHAR(255) NOT NULL,
			occurred_at VARCHAR(64) NOT NULL,
			description TEXT NOT NULL,
			contact VARCHAR(255) NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL,
			deleted_at VARCHAR(64) NULL,
			PRIMARY KEY (id),
			KEY idx_lost_items_deleted_at (deleted_at),
			KEY idx_lost_items_type_status (item_type, status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT ''`); err != nil {
		var me *mysql.MySQLError
		if !(errors.As(err, &me) && me.Number == 1060) {
			return err
		}
	}
	return nil
}
