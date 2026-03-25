package main

import (
	"database/sql"
	"log"
	"time"
)

func startBackgroundJobs(db *sql.DB) {
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
	if err := reconcileActivities(db, now); err != nil {
		log.Printf("reconcileActivities: %v", err)
	}
	if err := backfillActivityReminders(db); err != nil {
		log.Printf("backfillActivityReminders: %v", err)
	}
}

func backfillActivityReminders(db *sql.DB) error {
	rows, err := db.Query(
		`SELECT a.id, a.title, a.start_time, r.user_id
		 FROM activities a
		 JOIN activity_registrations r ON r.activity_id = a.id
		 WHERE a.deleted_at IS NULL AND a.status <> 'cancelled'`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var activityID, userID int64
		var title, startTime string
		if err := rows.Scan(&activityID, &title, &startTime, &userID); err != nil {
			return err
		}
		startAt, ok := parseFlexibleTime(startTime)
		if !ok {
			continue
		}
		if err := upsertActivityStartReminders(db, activityID, userID, title, startAt); err != nil {
			return err
		}
	}
	return rows.Err()
}

