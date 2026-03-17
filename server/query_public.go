package main

import (
	"database/sql"
	"strings"
)

func listActivities(db *sql.DB, keyword string, page, pageSize int) ([]Activity, int64, error) {
	where := ""
	args := []any{}
	if keyword != "" {
		where = "WHERE title LIKE ? OR summary LIKE ?"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	var total int64
	qCount := "SELECT COUNT(1) FROM activities " + where
	if err := db.QueryRow(qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := "SELECT id, title, cover_url, summary, content, location, start_time, end_time, created_at FROM activities " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Activity{}
	for rows.Next() {
		var it Activity
		if err := rows.Scan(&it.ID, &it.Title, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func getActivity(db *sql.DB, id int64) (Activity, error) {
	var it Activity
	err := db.QueryRow(
		"SELECT id, title, cover_url, summary, content, location, start_time, end_time, created_at FROM activities WHERE id = ?",
		id,
	).Scan(&it.ID, &it.Title, &it.CoverURL, &it.Summary, &it.Content, &it.Location, &it.StartTime, &it.EndTime, &it.CreatedAt)
	return it, err
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
