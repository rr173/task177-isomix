package store

import (
	"database/sql"
	"encoding/json"

	"task177-isomix/internal/model"
)

// EndmemberStore 端元持久化。
type EndmemberStore struct {
	db *sql.DB
}

// CreateEndmember 插入端元。
func (s *EndmemberStore) CreateEndmember(e *model.Endmember) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO endmembers (id, data, created_at) VALUES (?, ?, ?)`,
		e.ID, string(data), e.CreatedAt.UTC().Format(tsLayout))
	return err
}

// GetEndmember 按 ID 查询端元。
func (s *EndmemberStore) GetEndmember(id string) (*model.Endmember, error) {
	row := s.db.QueryRow(`SELECT data FROM endmembers WHERE id = ?`, id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var e model.Endmember
	if err := json.Unmarshal([]byte(data), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEndmember 更新端元（整体覆盖）。
func (s *EndmemberStore) UpdateEndmember(e *model.Endmember) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE endmembers SET data = ? WHERE id = ?`, string(data), e.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// ListEndmembers 列出全部端元。
func (s *EndmemberStore) ListEndmembers() ([]model.Endmember, error) {
	rows, err := s.db.Query(`SELECT data FROM endmembers ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Endmember
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var e model.Endmember
		if err := json.Unmarshal([]byte(data), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CountEndmembers 统计端元数。
func (s *EndmemberStore) CountEndmembers() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM endmembers`).Scan(&n)
	return n, err
}
