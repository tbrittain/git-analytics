package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"git-analytics/internal/git"
	"git-analytics/internal/store"
)

// sqliteStore implements store.Store using SQLite.
type sqliteStore struct {
	db     *sql.DB
	ownsDB bool
}

// Open opens or creates a SQLite database at the given path.
func Open(dbPath string) (store.Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	// WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	return &sqliteStore{db: db, ownsDB: true}, nil
}

// NewFromDB wraps an externally-owned *sql.DB. Close() is a no-op since the
// caller retains ownership of the database connection.
func NewFromDB(db *sql.DB) store.Store {
	return &sqliteStore{db: db, ownsDB: false}
}

func (s *sqliteStore) Init() error {
	_, err := s.db.Exec(store.SchemaSQL)
	return err
}

func (s *sqliteStore) InsertCommits(commits []git.Commit) error {
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

func (s *sqliteStore) GetLastIndexedCommit() (string, error) {
	var hash string
	err := s.db.QueryRow(
		`SELECT value FROM index_state WHERE key = 'last_indexed_commit'`).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return hash, err
}

func (s *sqliteStore) SetLastIndexedCommit(hash string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO index_state (key, value)
		 VALUES ('last_indexed_commit', ?)`, hash)
	return err
}

func (s *sqliteStore) Close() error {
	if s.ownsDB {
		return s.db.Close()
	}
	return nil
}
