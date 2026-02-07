package store

import "git-analytics/internal/git"

// Store persists extracted git analytics data.
type Store interface {
	// Init creates the database schema if it doesn't already exist.
	Init() error
	// InsertCommits inserts a batch of commits and their file stats.
	InsertCommits(commits []git.Commit) error
	// GetLastIndexedCommit returns the hash of the last indexed commit,
	// or an empty string if no commits have been indexed.
	GetLastIndexedCommit() (string, error)
	// SetLastIndexedCommit records the hash of the most recently indexed commit.
	SetLastIndexedCommit(hash string) error
	Close() error
}
