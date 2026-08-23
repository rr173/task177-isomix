package store

import (
	"database/sql"
	"encoding/json"

	"task177-isomix/internal/model"
)

// ReportStore 报告持久化。
type ReportStore struct {
	db *sql.DB
}

// CreateReport 插入报告。
func (s *ReportStore) CreateReport(r *model.Report) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO reports (id, data, solution_id, sample_id, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, string(data), r.SolutionID, r.SampleID, string(r.Status), r.CreatedAt.UTC().Format(tsLayout))
	return err
}

// UpdateReport 更新报告。
func (s *ReportStore) UpdateReport(r *model.Report) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE reports SET data = ?, solution_id = ?, sample_id = ?, status = ? WHERE id = ?`,
		string(data), r.SolutionID, r.SampleID, string(r.Status), r.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// GetReport 按 ID 查询报告。
func (s *ReportStore) GetReport(id string) (*model.Report, error) {
	row := s.db.QueryRow(`SELECT data FROM reports WHERE id = ?`, id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var r model.Report
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ListReports 列出全部报告。
func (s *ReportStore) ListReports() ([]model.Report, error) {
	rows, err := s.db.Query(`SELECT data FROM reports ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Report
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var r model.Report
		if err := json.Unmarshal([]byte(data), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CountReports 统计报告数。
func (s *ReportStore) CountReports() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM reports`).Scan(&n)
	return n, err
}
