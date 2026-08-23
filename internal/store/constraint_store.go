package store

import (
	"database/sql"
	"encoding/json"

	"task177-isomix/internal/model"
)

// ConstraintStore 约束持久化。
type ConstraintStore struct {
	db *sql.DB
}

// CreateConstraint 插入约束。
func (s *ConstraintStore) CreateConstraint(c *model.Constraint) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO constraints (id, data, created_at) VALUES (?, ?, ?)`,
		c.ID, string(data), c.CreatedAt.UTC().Format(tsLayout))
	return err
}

// GetConstraint 按 ID 查询约束。
func (s *ConstraintStore) GetConstraint(id string) (*model.Constraint, error) {
	row := s.db.QueryRow(`SELECT data FROM constraints WHERE id = ?`, id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var c model.Constraint
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateConstraint 更新约束。
func (s *ConstraintStore) UpdateConstraint(c *model.Constraint) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE constraints SET data = ? WHERE id = ?`, string(data), c.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// ListConstraints 列出全部约束。
func (s *ConstraintStore) ListConstraints() ([]model.Constraint, error) {
	rows, err := s.db.Query(`SELECT data FROM constraints ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Constraint
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var c model.Constraint
		if err := json.Unmarshal([]byte(data), &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CountConstraints 统计约束数。
func (s *ConstraintStore) CountConstraints() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM constraints`).Scan(&n)
	return n, err
}
