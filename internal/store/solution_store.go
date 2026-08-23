package store

import (
	"database/sql"
	"encoding/json"

	"task177-isomix/internal/model"
)

// SolutionStore 求解持久化。除 data 外另存 input_hash / sample_id / status 便于索引与恢复。
type SolutionStore struct {
	db *sql.DB
}

// CreateSolution 插入求解。
func (s *SolutionStore) CreateSolution(sol *model.Solution) error {
	data, err := json.Marshal(sol)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO solutions (id, data, input_hash, sample_id, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		sol.ID, string(data), sol.InputHash, sol.SampleID, string(sol.Status), sol.CreatedAt.UTC().Format(tsLayout))
	return err
}

// UpdateSolution 更新求解（同步更新索引列）。
func (s *SolutionStore) UpdateSolution(sol *model.Solution) error {
	data, err := json.Marshal(sol)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE solutions SET data = ?, input_hash = ?, sample_id = ?, status = ? WHERE id = ?`,
		string(data), sol.InputHash, sol.SampleID, string(sol.Status), sol.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// GetSolution 按 ID 查询求解。
func (s *SolutionStore) GetSolution(id string) (*model.Solution, error) {
	row := s.db.QueryRow(`SELECT data FROM solutions WHERE id = ?`, id)
	return scanSolution(row)
}

// GetSolutionByInputHash 按输入哈希查询求解（幂等复用）。
func (s *SolutionStore) GetSolutionByInputHash(h string) (*model.Solution, error) {
	row := s.db.QueryRow(`SELECT data FROM solutions WHERE input_hash = ? ORDER BY created_at LIMIT 1`, h)
	return scanSolution(row)
}

// ListSolutions 列出全部求解。
func (s *SolutionStore) ListSolutions() ([]model.Solution, error) {
	rows, err := s.db.Query(`SELECT data FROM solutions ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Solution
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var sol model.Solution
		if err := json.Unmarshal([]byte(data), &sol); err != nil {
			return nil, err
		}
		out = append(out, sol)
	}
	return out, rows.Err()
}

// CountSolutions 统计求解数。
func (s *SolutionStore) CountSolutions() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM solutions`).Scan(&n)
	return n, err
}

func scanSolution(row *sql.Row) (*model.Solution, error) {
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var sol model.Solution
	if err := json.Unmarshal([]byte(data), &sol); err != nil {
		return nil, err
	}
	return &sol, nil
}
