package main

import (
	"context"
	"log"
	"time"

	"community-help-hub-server/internal/config"
	"community-help-hub-server/internal/db"
	"community-help-hub-server/internal/httpapi"
	"community-help-hub-server/internal/jobs"
	"community-help-hub-server/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	port := config.EnvOr("PORT", "8080")
	origins := config.ParseOrigins(config.EnvOr("CORS_ORIGINS", ""))

	dsn, err := config.MySQLDSNFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := db.Open(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(pingCtx); err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(conn); err != nil {
		log.Fatal(err)
	}
	if config.AdminInitEnabled() {
		if err := store.EnsureAdminUser(conn); err != nil {
			log.Fatal(err)
		}
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(httpapi.WithCORS(origins))
	httpapi.RegisterRoutes(r, conn)
	jobs.StartBackgroundJobs(conn)

	addr := ":" + port
	log.Printf("server listening on %s", addr)
	log.Fatal(r.Run(addr))
}
