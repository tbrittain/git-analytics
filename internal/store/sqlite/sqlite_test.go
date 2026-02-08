package sqlite_test

import (
	"path/filepath"
	"testing"
	"time"

	"git-analytics/internal/git"
	sqlitestore "git-analytics/internal/store/sqlite"
)

func TestInitAndInsert(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := sqlitestore.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	commits := []git.Commit{
		{
			Hash:        "abc123def456abc123def456abc123def456abc1",
			AuthorName:  "Alice",
			AuthorEmail: "alice@example.com",
			Date:        time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			Message:     "initial commit",
			FilesChanged: []git.FileStat{
				{Path: "main.go", Additions: 50, Deletions: 0},
				{Path: "go.mod", Additions: 5, Deletions: 0},
			},
		},
		{
			Hash:        "def456abc123def456abc123def456abc123def4",
			AuthorName:  "Bob",
			AuthorEmail: "bob@example.com",
			Date:        time.Date(2025, 1, 16, 14, 0, 0, 0, time.UTC),
			Message:     "add feature",
			FilesChanged: []git.FileStat{
				{Path: "main.go", Additions: 10, Deletions: 3},
			},
		},
	}

	if err := s.InsertCommits(commits); err != nil {
		t.Fatalf("InsertCommits: %v", err)
	}

	// Inserting the same commits again should not fail (INSERT OR IGNORE).
	if err := s.InsertCommits(commits); err != nil {
		t.Fatalf("InsertCommits (duplicate): %v", err)
	}
}

func TestIndexState(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := sqlitestore.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Initially, no last indexed commit.
	hash, err := s.GetLastIndexedCommit()
	if err != nil {
		t.Fatalf("GetLastIndexedCommit: %v", err)
	}
	if hash != "" {
		t.Errorf("expected empty hash, got %q", hash)
	}

	// Set and read back.
	if err := s.SetLastIndexedCommit("abc123"); err != nil {
		t.Fatalf("SetLastIndexedCommit: %v", err)
	}

	hash, err = s.GetLastIndexedCommit()
	if err != nil {
		t.Fatalf("GetLastIndexedCommit: %v", err)
	}
	if hash != "abc123" {
		t.Errorf("expected 'abc123', got %q", hash)
	}

	// Update.
	if err := s.SetLastIndexedCommit("def456"); err != nil {
		t.Fatalf("SetLastIndexedCommit: %v", err)
	}

	hash, err = s.GetLastIndexedCommit()
	if err != nil {
		t.Fatalf("GetLastIndexedCommit: %v", err)
	}
	if hash != "def456" {
		t.Errorf("expected 'def456', got %q", hash)
	}
}

func TestInitIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := sqlitestore.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Init twice should not fail.
	if err := s.Init(); err != nil {
		t.Fatalf("Init (first): %v", err)
	}
	if err := s.Init(); err != nil {
		t.Fatalf("Init (second): %v", err)
	}
}
