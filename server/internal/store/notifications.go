package store

import (
	"database/sql"
	"time"

	"community-help-hub-server/internal/domain"
)

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

func ListUserNotifications(db *sql.DB, userID int64, page, pageSize int) ([]domain.Notification, int64, error) {
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
	out := []domain.Notification{}
	for rows.Next() {
		var it domain.Notification
		if err := rows.Scan(&it.ID, &it.Kind, &it.Title, &it.Content, &it.ActivityID, &it.ScheduledFor, &it.ReadAt, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func MarkNotificationRead(db *sql.DB, userID, notificationID int64) error {
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

func BackfillActivityReminders(db *sql.DB) error {
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
