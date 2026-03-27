package jobs

import (
	"database/sql"
	"log"
	"time"

	"community-help-hub-server/internal/store"
)

func StartBackgroundJobs(db *sql.DB) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			runBackgroundJobsOnce(db)
			<-ticker.C
		}
	}()
}

func runBackgroundJobsOnce(db *sql.DB) {
	now := time.Now()
	if err := store.ReconcileActivities(db, now); err != nil {
		log.Printf("reconcileActivities: %v", err)
	}
	if err := store.BackfillActivityReminders(db); err != nil {
		log.Printf("backfillActivityReminders: %v", err)
	}
}
