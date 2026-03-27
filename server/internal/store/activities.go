package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"community-help-hub-server/internal/domain"

	mysql "github.com/go-sql-driver/mysql"
)

func parseFlexibleTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func deriveActivityStatus(now time.Time, status, startTime, endTime string) (string, time.Time, bool) {
	if status == "cancelled" {
		return status, time.Time{}, false
	}
	startAt, okStart := parseFlexibleTime(startTime)
	endAt, okEnd := parseFlexibleTime(endTime)
	if okEnd && (now.Equal(endAt) || now.After(endAt)) {
		return "finished", time.Time{}, okStart
	}
	if okStart {
		deadline := startAt.Add(-24 * time.Hour)
		if now.Before(deadline) {
			return "active", deadline, true
		}
		return "closed", deadline, true
	}
	return status, time.Time{}, false
}

func shouldUpdateActivityStatus(oldStatus, newStatus string) bool {
	if oldStatus == "cancelled" {
		return false
	}
	return oldStatus != newStatus
}

func ReconcileActivityStatusByID(db *sql.DB, id int64, now time.Time) (string, error) {
	var status, startTime, endTime string
	err := db.QueryRow(
		"SELECT status, start_time, end_time FROM activities WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&status, &startTime, &endTime)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrActivityNotFound
	}
	if err != nil {
		return "", err
	}
	next, _, _ := deriveActivityStatus(now, status, startTime, endTime)
	if shouldUpdateActivityStatus(status, next) && (next == "active" || next == "closed" || next == "finished") {
		if _, err := db.Exec("UPDATE activities SET status = ? WHERE id = ? AND status <> 'cancelled'", next, id); err != nil {
			return "", err
		}
	}
	return next, nil
}

func ReconcileActivities(db *sql.DB, now time.Time) error {
	rows, err := db.Query("SELECT id, status, start_time, end_time FROM activities WHERE deleted_at IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var status, startTime, endTime string
		if err := rows.Scan(&id, &status, &startTime, &endTime); err != nil {
			return err
		}
		next, _, _ := deriveActivityStatus(now, status, startTime, endTime)
		if shouldUpdateActivityStatus(status, next) && (next == "active" || next == "closed" || next == "finished") {
			if _, err := db.Exec("UPDATE activities SET status = ? WHERE id = ? AND status <> 'cancelled'", next, id); err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func ListActivities(db *sql.DB, category, status, keyword string, page, pageSize int) ([]domain.Activity, int64, error) {
	clauses := []string{}
	args := []any{}
	clauses = append(clauses, "deleted_at IS NULL")
	if category != "" {
		clauses = append(clauses, "category = ?")
		args = append(args, category)
	}
	if status != "" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if keyword != "" {
		clauses = append(clauses, "(title LIKE ? OR summary LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	qCount := "SELECT COUNT(1) FROM activities " + where
	if err := db.QueryRow(qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at FROM activities " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Activity{}
	for rows.Next() {
		var it domain.Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func GetActivity(db *sql.DB, id int64) (domain.Activity, error) {
	var it domain.Activity
	err := db.QueryRow(
		"SELECT id, title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at FROM activities WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt)
	return it, err
}

func RegisterActivity(db *sql.DB, activityID, userID int64) error {
	now := time.Now()
	var status, startTime, endTime, title string
	err := db.QueryRow(
		"SELECT status, start_time, end_time, title FROM activities WHERE id = ? AND deleted_at IS NULL",
		activityID,
	).Scan(&status, &startTime, &endTime, &title)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrActivityNotFound
	}
	if err != nil {
		return err
	}
	next, deadline, okDeadline := deriveActivityStatus(now, status, startTime, endTime)
	if shouldUpdateActivityStatus(status, next) && (next == "active" || next == "closed" || next == "finished") {
		if _, err := db.Exec("UPDATE activities SET status = ? WHERE id = ? AND status <> 'cancelled'", next, activityID); err != nil {
			return err
		}
	}
	if next != "active" {
		return ErrRegistrationClosed
	}
	if !okDeadline {
		return ErrActivityTimeInvalid
	}
	if now.Equal(deadline) || now.After(deadline) {
		return ErrRegistrationClosed
	}

	createdAt := now.Format(time.RFC3339)
	_, err = db.Exec(
		"INSERT INTO activity_registrations (activity_id, user_id, status, created_at) VALUES (?, ?, 'pending', ?)",
		activityID, userID, createdAt,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			var existing string
			if qErr := db.QueryRow(
				"SELECT status FROM activity_registrations WHERE activity_id = ? AND user_id = ?",
				activityID, userID,
			).Scan(&existing); qErr != nil {
				return ErrAlreadyRegistered
			}
			if existing == "cancelled" {
				if _, uErr := db.Exec(
					"UPDATE activity_registrations SET status = 'pending', created_at = ? WHERE activity_id = ? AND user_id = ?",
					createdAt, activityID, userID,
				); uErr != nil {
					return uErr
				}
				if startAt, ok := parseFlexibleTime(startTime); ok {
					_ = upsertActivityStartReminders(db, activityID, userID, title, startAt)
				}
				return nil
			}
			return ErrAlreadyRegistered
		}
		return err
	}
	if startAt, ok := parseFlexibleTime(startTime); ok {
		_ = upsertActivityStartReminders(db, activityID, userID, title, startAt)
	}
	return nil
}

func CancelActivityRegistration(db *sql.DB, activityID, userID int64) error {
	now := time.Now()
	var status, startTime, endTime string
	err := db.QueryRow(
		"SELECT status, start_time, end_time FROM activities WHERE id = ? AND deleted_at IS NULL",
		activityID,
	).Scan(&status, &startTime, &endTime)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrActivityNotFound
	}
	if err != nil {
		return err
	}

	next, deadline, okDeadline := deriveActivityStatus(now, status, startTime, endTime)
	if shouldUpdateActivityStatus(status, next) && (next == "active" || next == "closed" || next == "finished") {
		if _, err := db.Exec("UPDATE activities SET status = ? WHERE id = ? AND status <> 'cancelled'", next, activityID); err != nil {
			return err
		}
	}
	if next != "active" {
		return ErrCancellationClosed
	}
	if !okDeadline {
		return ErrActivityTimeInvalid
	}
	if now.Equal(deadline) || now.After(deadline) {
		return ErrCancellationClosed
	}

	res, err := db.Exec(
		"UPDATE activity_registrations SET status = 'cancelled' WHERE activity_id = ? AND user_id = ? AND status <> 'cancelled'",
		activityID, userID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotRegistered
	}
	return nil
}

func ListUserRegisteredActivities(db *sql.DB, userID int64) ([]domain.Activity, error) {
	q := `SELECT a.id, a.title, a.category, a.status, a.user_id, a.cover_url, a.summary, a.content, a.location, a.start_time, a.end_time, a.created_at 
		  FROM activities a 
		  JOIN activity_registrations r ON a.id = r.activity_id 
		  WHERE r.user_id = ? AND r.status <> 'cancelled' AND a.deleted_at IS NULL ORDER BY r.created_at DESC`
	rows, err := db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Activity{}
	for rows.Next() {
		var it domain.Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func ListUserPublishedActivities(db *sql.DB, userID int64) ([]domain.Activity, error) {
	q := "SELECT id, title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at FROM activities WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC"
	rows, err := db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Activity{}
	for rows.Next() {
		var it domain.Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func CreateActivity(db *sql.DB, it domain.Activity) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	res, err := db.Exec(
		"INSERT INTO activities (title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		it.Title, it.Category, it.Status, it.UserID, it.CoverURL, it.Summary, it.Content, it.Location, it.StartTime, it.EndTime, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateActivityByOwner(db *sql.DB, id, ownerUserID int64, it domain.Activity) error {
	var currentOwner int64
	err := db.QueryRow("SELECT user_id FROM activities WHERE id = ? AND deleted_at IS NULL", id).Scan(&currentOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrActivityNotFound
	}
	if err != nil {
		return err
	}
	if currentOwner != ownerUserID {
		return ErrForbidden
	}

	nextStatus := strings.TrimSpace(it.Status)
	if nextStatus != "cancelled" {
		nextStatus = ""
	}
	if nextStatus == "" {
		if err := db.QueryRow("SELECT status FROM activities WHERE id = ? AND deleted_at IS NULL", id).Scan(&nextStatus); err != nil {
			return err
		}
	}

	_, err = db.Exec(
		"UPDATE activities SET title = ?, category = ?, status = ?, cover_url = ?, summary = ?, content = ?, location = ?, start_time = ?, end_time = ? WHERE id = ?",
		it.Title, it.Category, nextStatus, it.CoverURL, it.Summary, it.Content, it.Location, it.StartTime, it.EndTime, id,
	)
	if err != nil {
		return err
	}
	if nextStatus == "cancelled" {
		return deleteNotificationsByActivity(db, id)
	}
	if startAt, ok := parseFlexibleTime(it.StartTime); ok {
		return upsertActivityStartRemindersForActivity(db, id, it.Title, startAt)
	}
	return nil
}

func DeleteActivityByOwner(db *sql.DB, id, ownerUserID int64) error {
	var currentOwner int64
	err := db.QueryRow("SELECT user_id FROM activities WHERE id = ? AND deleted_at IS NULL", id).Scan(&currentOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrActivityNotFound
	}
	if err != nil {
		return err
	}
	if currentOwner != ownerUserID {
		return ErrForbidden
	}
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec("UPDATE activities SET deleted_at = ? WHERE id = ?", now, id)
	if err != nil {
		return err
	}
	return deleteNotificationsByActivity(db, id)
}

func ListActivityRegistrationsByOwner(db *sql.DB, activityID, ownerUserID int64) ([]domain.ActivityRegistrationView, error) {
	var currentOwner int64
	err := db.QueryRow("SELECT user_id FROM activities WHERE id = ? AND deleted_at IS NULL", activityID).Scan(&currentOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrActivityNotFound
	}
	if err != nil {
		return nil, err
	}
	if currentOwner != ownerUserID {
		return nil, ErrForbidden
	}

	rows, err := db.Query(
		`SELECT r.id, r.activity_id, r.user_id, u.username, r.status, r.created_at
		 FROM activity_registrations r
		 JOIN users u ON u.id = r.user_id
		 WHERE r.activity_id = ?
		 ORDER BY r.created_at DESC`,
		activityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ActivityRegistrationView{}
	for rows.Next() {
		var it domain.ActivityRegistrationView
		if err := rows.Scan(&it.ID, &it.ActivityID, &it.UserID, &it.Username, &it.Status, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
