package store

import (
	"database/sql"
	"strings"
	"time"

	"community-help-hub-server/internal/domain"
)

func ListServices(db *sql.DB, category, keyword string, page, pageSize int) ([]domain.Service, int64, error) {
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
	out := []domain.Service{}
	for rows.Next() {
		var it domain.Service
		if err := rows.Scan(&it.ID, &it.Name, &it.Category, &it.Phone, &it.Address, &it.Description, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func GetService(db *sql.DB, id int64) (domain.Service, error) {
	var it domain.Service
	err := db.QueryRow(
		"SELECT id, name, category, phone, address, description, updated_at FROM services WHERE id = ?",
		id,
	).Scan(&it.ID, &it.Name, &it.Category, &it.Phone, &it.Address, &it.Description, &it.UpdatedAt)
	return it, err
}

func CreateService(db *sql.DB, it domain.Service) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	res, err := db.Exec(
		"INSERT INTO services (name, category, phone, address, description, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		it.Name, it.Category, it.Phone, it.Address, it.Description, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateService(db *sql.DB, id int64, it domain.Service) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec(
		"UPDATE services SET name = ?, category = ?, phone = ?, address = ?, description = ?, updated_at = ? WHERE id = ?",
		it.Name, it.Category, it.Phone, it.Address, it.Description, now, id,
	)
	return err
}

func DeleteService(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM services WHERE id = ?", id)
	return err
}
