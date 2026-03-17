package main

import (
	"database/sql"
	"strings"
	"time"
)

func createLostItem(db *sql.DB, it LostItem) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	if it.Status == "" {
		it.Status = "open"
	}
	res, err := db.Exec(
		`INSERT INTO lost_items (title, item_type, status, location, occurred_at, description, contact, created_at, updated_at) 
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(it.Title),
		strings.TrimSpace(it.ItemType),
		strings.TrimSpace(it.Status),
		strings.TrimSpace(it.Location),
		strings.TrimSpace(it.OccurredAt),
		strings.TrimSpace(it.Description),
		strings.TrimSpace(it.Contact),
		now,
		now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateLostItem(db *sql.DB, id int64, it LostItem) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec(
		`UPDATE lost_items 
		 SET title = ?, item_type = ?, status = ?, location = ?, occurred_at = ?, description = ?, contact = ?, updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		strings.TrimSpace(it.Title),
		strings.TrimSpace(it.ItemType),
		strings.TrimSpace(it.Status),
		strings.TrimSpace(it.Location),
		strings.TrimSpace(it.OccurredAt),
		strings.TrimSpace(it.Description),
		strings.TrimSpace(it.Contact),
		now,
		id,
	)
	return err
}

func deleteLostItem(db *sql.DB, id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec("UPDATE lost_items SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", now, now, id)
	return err
}
