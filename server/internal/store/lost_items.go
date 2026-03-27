package store

import (
	"database/sql"
	"strings"
	"time"

	"community-help-hub-server/internal/domain"
)

func CreateLostItem(db *sql.DB, it domain.LostItem) (int64, error) {
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

func UpdateLostItem(db *sql.DB, id int64, it domain.LostItem) error {
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

func DeleteLostItem(db *sql.DB, id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec("UPDATE lost_items SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", now, now, id)
	return err
}

func ListLostItems(db *sql.DB, itemType, status, keyword string, page, pageSize int, includeDeleted bool) ([]domain.LostItem, int64, error) {
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
	out := []domain.LostItem{}
	for rows.Next() {
		var it domain.LostItem
		if err := rows.Scan(&it.ID, &it.Title, &it.ItemType, &it.Status, &it.Location, &it.OccurredAt, &it.Description, &it.Contact, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func GetLostItem(db *sql.DB, id int64) (domain.LostItem, error) {
	var it domain.LostItem
	err := db.QueryRow(
		"SELECT id, title, item_type, status, location, occurred_at, description, contact, created_at, updated_at FROM lost_items WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&it.ID, &it.Title, &it.ItemType, &it.Status, &it.Location, &it.OccurredAt, &it.Description, &it.Contact, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}
