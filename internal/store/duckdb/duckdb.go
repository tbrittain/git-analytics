package duckdb

import (
	"database/sql"

	_ "github.com/duckdb/duckdb-go/v2"

	"git-analytics/internal/git"
	"git-analytics/internal/store"
)

// duckDBStore implements store.Store using DuckDB.
type duckDBStore struct {
	db *sql.DB
}

// Open opens or creates a DuckDB database at the given path.
func Open(dbPath string) (store.Store, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, err
	}
	return &duckDBStore{db: db}, nil
}

func (s *duckDBStore) Init() error {
	_, err := s.db.Exec(store.SchemaSQL)
	return err
}

func (s *duckDBStore) InsertCommits(commits []git.Commit) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	commitStmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO commits (hash, author_name, author_email, committed_at, message)
		 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer commitStmt.Close()

	fileStmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO file_stats (commit_hash, file_path, additions, deletions)
		 VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer fileStmt.Close()

	for _, c := range commits {
		_, err := commitStmt.Exec(c.Hash, c.AuthorName, c.AuthorEmail, c.Date, c.Message)
		if err != nil {
			return err
		}
		for _, f := range c.FilesChanged {
			_, err := fileStmt.Exec(c.Hash, f.Path, f.Additions, f.Deletions)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *duckDBStore) GetLastIndexedCommit() (string, error) {
	var hash string
	err := s.db.QueryRow(
		`SELECT value FROM index_state WHERE key = 'last_indexed_commit'`).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return hash, err
}

func (s *duckDBStore) SetLastIndexedCommit(hash string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO index_state (key, value)
		 VALUES ('last_indexed_commit', ?)`, hash)
	return err
}

func (s *duckDBStore) Close() error {
	return s.db.Close()
}
