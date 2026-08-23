package store

import (
	"database/sql"
	"encoding/json"

	"task177-isomix/internal/model"
)

// SampleStore 样品持久化。
type SampleStore struct {
	db *sql.DB
}

// CreateSample 插入样品。
func (s *SampleStore) CreateSample(sp *model.Sample) error {
	data, err := json.Marshal(sp)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO samples (id, data, created_at) VALUES (?, ?, ?)`,
		sp.ID, string(data), sp.CreatedAt.UTC().Format(tsLayout))
	return err
}

// GetSample 按 ID 查询样品。
func (s *SampleStore) GetSample(id string) (*model.Sample, error) {
	row := s.db.QueryRow(`SELECT data FROM samples WHERE id = ?`, id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var sp model.Sample
	if err := json.Unmarshal([]byte(data), &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// UpdateSample 更新样品。
func (s *SampleStore) UpdateSample(sp *model.Sample) error {
	data, err := json.Marshal(sp)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE samples SET data = ? WHERE id = ?`, string(data), sp.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// ListSamples 列出全部样品。
func (s *SampleStore) ListSamples() ([]model.Sample, error) {
	rows, err := s.db.Query(`SELECT data FROM samples ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Sample
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var sp model.Sample
		if err := json.Unmarshal([]byte(data), &sp); err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

// CountSamples 统计样品数。
func (s *SampleStore) CountSamples() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM samples`).Scan(&n)
	return n, err
}
