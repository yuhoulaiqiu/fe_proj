package main

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	errActivityNotFound    = errors.New("activity_not_found")
	errRegistrationClosed  = errors.New("registration_closed")
	errCancellationClosed  = errors.New("cancellation_closed")
	errAlreadyRegistered   = errors.New("already_registered")
	errNotRegistered       = errors.New("not_registered")
	errActivityTimeInvalid = errors.New("activity_time_invalid")
	errForbidden           = errors.New("forbidden")
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

func reconcileActivityStatusByID(db *sql.DB, id int64, now time.Time) (string, error) {
	var status, startTime, endTime string
	err := db.QueryRow(
		"SELECT status, start_time, end_time FROM activities WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&status, &startTime, &endTime)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errActivityNotFound
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

func reconcileActivities(db *sql.DB, now time.Time) error {
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

func listActivities(db *sql.DB, category, status, keyword string, page, pageSize int) ([]Activity, int64, error) {
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
	out := []Activity{}
	for rows.Next() {
		var it Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getActivity(db *sql.DB, id int64) (Activity, error) {
	var it Activity
	err := db.QueryRow(
		"SELECT id, title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at FROM activities WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt)
	return it, err
}

func registerActivity(db *sql.DB, activityID, userID int64) error {
	now := time.Now()
	var status, startTime, endTime, title string
	err := db.QueryRow(
		"SELECT status, start_time, end_time, title FROM activities WHERE id = ? AND deleted_at IS NULL",
		activityID,
	).Scan(&status, &startTime, &endTime, &title)
	if errors.Is(err, sql.ErrNoRows) {
		return errActivityNotFound
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
		return errRegistrationClosed
	}
	if !okDeadline {
		return errActivityTimeInvalid
	}
	if now.Equal(deadline) || now.After(deadline) {
		return errRegistrationClosed
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
				return errAlreadyRegistered
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
			return errAlreadyRegistered
		}
		return err
	}
	if startAt, ok := parseFlexibleTime(startTime); ok {
		_ = upsertActivityStartReminders(db, activityID, userID, title, startAt)
	}
	return nil
}

func cancelActivityRegistration(db *sql.DB, activityID, userID int64) error {
	now := time.Now()
	var status, startTime, endTime string
	err := db.QueryRow(
		"SELECT status, start_time, end_time FROM activities WHERE id = ? AND deleted_at IS NULL",
		activityID,
	).Scan(&status, &startTime, &endTime)
	if errors.Is(err, sql.ErrNoRows) {
		return errActivityNotFound
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
		return errCancellationClosed
	}
	if !okDeadline {
		return errActivityTimeInvalid
	}
	if now.Equal(deadline) || now.After(deadline) {
		return errCancellationClosed
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
		return errNotRegistered
	}
	return nil
}

func listUserRegisteredActivities(db *sql.DB, userID int64) ([]Activity, error) {
	q := `SELECT a.id, a.title, a.category, a.status, a.user_id, a.cover_url, a.summary, a.content, a.location, a.start_time, a.end_time, a.created_at 
		  FROM activities a 
		  JOIN activity_registrations r ON a.id = r.activity_id 
		  WHERE r.user_id = ? AND r.status <> 'cancelled' AND a.deleted_at IS NULL ORDER BY r.created_at DESC`
	rows, err := db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Activity{}
	for rows.Next() {
		var it Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func listUserPublishedActivities(db *sql.DB, userID int64) ([]Activity, error) {
	q := "SELECT id, title, category, status, user_id, cover_url, summary, content, location, start_time, end_time, created_at FROM activities WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC"
	rows, err := db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Activity{}
	for rows.Next() {
		var it Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.Status, &it.UserID, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func createActivity(db *sql.DB, it Activity) (int64, error) {
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

func updateActivityByOwner(db *sql.DB, id, ownerUserID int64, it Activity) error {
	var currentOwner int64
	err := db.QueryRow("SELECT user_id FROM activities WHERE id = ? AND deleted_at IS NULL", id).Scan(&currentOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return errActivityNotFound
	}
	if err != nil {
		return err
	}
	if currentOwner != ownerUserID {
		return errForbidden
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

func deleteActivityByOwner(db *sql.DB, id, ownerUserID int64) error {
	var currentOwner int64
	err := db.QueryRow("SELECT user_id FROM activities WHERE id = ? AND deleted_at IS NULL", id).Scan(&currentOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return errActivityNotFound
	}
	if err != nil {
		return err
	}
	if currentOwner != ownerUserID {
		return errForbidden
	}
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec("UPDATE activities SET deleted_at = ? WHERE id = ?", now, id)
	if err != nil {
		return err
	}
	return deleteNotificationsByActivity(db, id)
}

func listActivityRegistrationsByOwner(db *sql.DB, activityID, ownerUserID int64) ([]ActivityRegistrationView, error) {
	var currentOwner int64
	err := db.QueryRow("SELECT user_id FROM activities WHERE id = ? AND deleted_at IS NULL", activityID).Scan(&currentOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errActivityNotFound
	}
	if err != nil {
		return nil, err
	}
	if currentOwner != ownerUserID {
		return nil, errForbidden
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
	out := []ActivityRegistrationView{}
	for rows.Next() {
		var it ActivityRegistrationView
		if err := rows.Scan(&it.ID, &it.ActivityID, &it.UserID, &it.Username, &it.Status, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func upsertNotification(db *sql.DB, userID int64, kind, title, content string, activityID int64, scheduledFor time.Time) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec(
		`INSERT INTO notifications (user_id, kind, title, content, activity_id, scheduled_for, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE title = VALUES(title), content = VALUES(content), scheduled_for = VALUES(scheduled_for)`,
		userID, kind, title, content, activityID, scheduledFor.Format(time.RFC3339), now,
	)
	return err
}

func listUserNotifications(db *sql.DB, userID int64, page, pageSize int) ([]Notification, int64, error) {
	now := time.Now().Format(time.RFC3339)
	var total int64
	if err := db.QueryRow(
		"SELECT COUNT(1) FROM notifications WHERE user_id = ? AND scheduled_for <= ?",
		userID, now,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := db.Query(
		`SELECT id, kind, title, content, IFNULL(activity_id, 0), scheduled_for, IFNULL(read_at, ''), created_at
		 FROM notifications
		 WHERE user_id = ? AND scheduled_for <= ?
		 ORDER BY scheduled_for DESC, id DESC
		 LIMIT ? OFFSET ?`,
		userID, now, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Notification{}
	for rows.Next() {
		var it Notification
		if err := rows.Scan(&it.ID, &it.Kind, &it.Title, &it.Content, &it.ActivityID, &it.ScheduledFor, &it.ReadAt, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func markNotificationRead(db *sql.DB, userID, notificationID int64) error {
	now := time.Now().Format(time.RFC3339)
	res, err := db.Exec(
		"UPDATE notifications SET read_at = ? WHERE id = ? AND user_id = ? AND read_at IS NULL",
		now, notificationID, userID,
	)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func deleteNotificationsByActivity(db *sql.DB, activityID int64) error {
	_, err := db.Exec("DELETE FROM notifications WHERE activity_id = ?", activityID)
	return err
}

func listActivityRegistrationUserIDs(db *sql.DB, activityID int64) ([]int64, error) {
	rows, err := db.Query("SELECT user_id FROM activity_registrations WHERE activity_id = ? ORDER BY id DESC", activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []int64{}
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		out = append(out, uid)
	}
	return out, rows.Err()
}

func upsertActivityStartReminders(db *sql.DB, activityID, userID int64, activityTitle string, startAt time.Time) error {
	t24 := startAt.Add(-24 * time.Hour)
	t1 := startAt.Add(-1 * time.Hour)

	title24 := "活动提醒"
	content24 := activityTitle + " 将于 " + startAt.Format(time.RFC3339) + " 开始（提前 24 小时提醒）"
	title1 := "活动提醒"
	content1 := activityTitle + " 将于 " + startAt.Format(time.RFC3339) + " 开始（提前 1 小时提醒）"

	if err := upsertNotification(db, userID, "activity_start_24h", title24, content24, activityID, t24); err != nil {
		return err
	}
	if err := upsertNotification(db, userID, "activity_start_1h", title1, content1, activityID, t1); err != nil {
		return err
	}
	return nil
}

func upsertActivityStartRemindersForActivity(db *sql.DB, activityID int64, activityTitle string, startAt time.Time) error {
	userIDs, err := listActivityRegistrationUserIDs(db, activityID)
	if err != nil {
		return err
	}
	for _, uid := range userIDs {
		if err := upsertActivityStartReminders(db, activityID, uid, activityTitle, startAt); err != nil {
			return err
		}
	}
	return nil
}

func listServices(db *sql.DB, category, keyword string, page, pageSize int) ([]Service, int64, error) {
	clauses := []string{}
	args := []any{}
	if category != "" {
		clauses = append(clauses, "category = ?")
		args = append(args, category)
	}
	if keyword != "" {
		clauses = append(clauses, "(name LIKE ? OR description LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := db.QueryRow("SELECT COUNT(1) FROM services "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, name, category, phone, address, description, updated_at FROM services " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Service{}
	for rows.Next() {
		var it Service
		if err := rows.Scan(&it.ID, &it.Name, &it.Category, &it.Phone, &it.Address, &it.Description, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getService(db *sql.DB, id int64) (Service, error) {
	var it Service
	err := db.QueryRow(
		"SELECT id, name, category, phone, address, description, updated_at FROM services WHERE id = ?",
		id,
	).Scan(&it.ID, &it.Name, &it.Category, &it.Phone, &it.Address, &it.Description, &it.UpdatedAt)
	return it, err
}

func listLostItems(db *sql.DB, itemType, status, keyword string, page, pageSize int, includeDeleted bool) ([]LostItem, int64, error) {
	clauses := []string{}
	args := []any{}
	if !includeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	if itemType != "" {
		clauses = append(clauses, "item_type = ?")
		args = append(args, itemType)
	}
	if status != "" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if keyword != "" {
		clauses = append(clauses, "(title LIKE ? OR description LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := db.QueryRow("SELECT COUNT(1) FROM lost_items "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, title, item_type, status, location, occurred_at, description, contact, created_at, updated_at FROM lost_items " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []LostItem{}
	for rows.Next() {
		var it LostItem
		if err := rows.Scan(&it.ID, &it.Title, &it.ItemType, &it.Status, &it.Location, &it.OccurredAt, &it.Description, &it.Contact, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getLostItem(db *sql.DB, id int64) (LostItem, error) {
	var it LostItem
	err := db.QueryRow(
		"SELECT id, title, item_type, status, location, occurred_at, description, contact, created_at, updated_at FROM lost_items WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&it.ID, &it.Title, &it.ItemType, &it.Status, &it.Location, &it.OccurredAt, &it.Description, &it.Contact, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}
