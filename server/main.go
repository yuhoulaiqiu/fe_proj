package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	port := envOr("PORT", "8080")
	origins := parseOrigins(os.Getenv("CORS_ORIGINS"))

	dsn, err := mysqlDSNFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		log.Fatal(err)
	}

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}
	if adminInitEnabled() {
		if err := ensureAdminUser(db); err != nil {
			log.Fatal(err)
		}
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(withCORS(origins))
	registerRoutes(r, db)
	startBackgroundJobs(db)

	addr := ":" + port
	log.Printf("server listening on %s", addr)
	log.Fatal(r.Run(addr))
}
