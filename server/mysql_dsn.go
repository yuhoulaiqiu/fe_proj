package main

import (
	"os"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

func mysqlDSNFromEnv() (string, error) {
	if v := strings.TrimSpace(os.Getenv("MYSQL_DSN")); v != "" {
		return v, nil
	}
	host := envOr("MYSQL_HOST", "127.0.0.1")
	port := envOr("MYSQL_PORT", "3306")
	user := envOr("MYSQL_USER", "root")
	pass := os.Getenv("MYSQL_PASSWORD")
	dbName := envOr("MYSQL_DATABASE", "community_help_hub")

	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Net = "tcp"
	cfg.Addr = host + ":" + port
	cfg.DBName = dbName
	cfg.Collation = "utf8mb4_unicode_ci"
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Timeout = 5 * time.Second
	cfg.ReadTimeout = 10 * time.Second
	cfg.WriteTimeout = 10 * time.Second
	return cfg.FormatDSN(), nil
}
