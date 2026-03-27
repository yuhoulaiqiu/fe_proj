package config

import (
	"os"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

func EnvOr(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func ParseOrigins(s string) []string {
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

func AdminInitEnabled() bool {
	v := strings.TrimSpace(os.Getenv("ADMIN_INIT"))
	if v == "" {
		return true
	}
	v = strings.ToLower(v)
	return !(v == "0" || v == "false" || v == "off" || v == "no")
}

func MySQLDSNFromEnv() (string, error) {
	if v := strings.TrimSpace(os.Getenv("MYSQL_DSN")); v != "" {
		return v, nil
	}
	host := EnvOr("MYSQL_HOST", "127.0.0.1")
	port := EnvOr("MYSQL_PORT", "3306")
	user := EnvOr("MYSQL_USER", "root")
	pass := os.Getenv("MYSQL_PASSWORD")
	dbName := EnvOr("MYSQL_DATABASE", "community_help_hub")

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
