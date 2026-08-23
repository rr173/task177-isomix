// Package store 提供 SQLite 持久化：建表迁移、各实体 JSON 存取与索引。
// 使用纯 Go 驱动 modernc.org/sqlite（CGO 无关，离线可构建）。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store 持有数据库连接并组合各实体存储。
type Store struct {
	db *sql.DB
	*EndmemberStore
	*SampleStore
	*ConstraintStore
	*SolutionStore
	*ReportStore
}

// Open 打开（或创建）SQLite 数据库并执行迁移。
// dbPath 为空时使用内存库（:memory:，供测试与自检）。
func Open(dbPath string) (*Store, error) {
	if dbPath != "" && dbPath != ":memory:" {
		if dir := filepath.Dir(dbPath); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create db dir: %w", err)
			}
		}
	}
	dsn := dbPath
	if dsn == "" {
		dsn = ":memory:"
	} else {
		// WAL 模式提升并发读写的健壮性；foreign_keys 保持默认关闭（本服务无外键约束）。
		dsn = dsn + "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	s.EndmemberStore = &EndmemberStore{db: db}
	s.SampleStore = &SampleStore{db: db}
	s.ConstraintStore = &ConstraintStore{db: db}
	s.SolutionStore = &SolutionStore{db: db}
	s.ReportStore = &ReportStore{db: db}
	return s, nil
}

// migrate 幂等建表。
func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS endmembers (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS samples (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS constraints (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS solutions (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			sample_id TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_solutions_hash ON solutions(input_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_solutions_sample ON solutions(sample_id)`,
		`CREATE INDEX IF NOT EXISTS idx_solutions_status ON solutions(status)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			solution_id TEXT NOT NULL,
			sample_id TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_solution ON reports(solution_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// DB 返回底层连接（仅供自检/统计使用）。
func (s *Store) DB() *sql.DB { return s.db }
