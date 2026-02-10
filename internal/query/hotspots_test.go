package query_test

import (
	"database/sql"
	"testing"
	"time"

	"git-analytics/internal/query"
)

func insertFileStat(t *testing.T, db *sql.DB, commitHash, filePath string, additions, deletions int) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO file_stats (commit_hash, file_path, additions, deletions) VALUES (?, ?, ?, ?)`,
		commitHash, filePath, additions, deletions,
	)
	if err != nil {
		t.Fatalf("insert file_stat: %v", err)
	}
}

func TestFileHotspots_Basic(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "aaa2", "Alice", "alice@example.com",
		time.Date(2025, 1, 16, 12, 0, 0, 0, time.UTC), "second")

	// commit aaa1 touches main.go and util.go
	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "aaa1", "util.go", 3, 2)
	// commit aaa2 touches main.go and readme.md
	insertFileStat(t, db, "aaa2", "main.go", 20, 10)
	insertFileStat(t, db, "aaa2", "readme.md", 1, 0)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.FileHotspots(db, from, to)
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 3 {
		t.Fatalf("expected 3 hotspots, got %d: %v", len(hotspots), hotspots)
	}

	// main.go: (10+5) + (20+10) = 45 lines, 2 commits
	if hotspots[0].Path != "main.go" || hotspots[0].LinesChanged != 45 || hotspots[0].Commits != 2 {
		t.Errorf("hotspot 0: got %+v, want {main.go, 45, 2}", hotspots[0])
	}
	// util.go: 3+2 = 5 lines, 1 commit
	if hotspots[1].Path != "util.go" || hotspots[1].LinesChanged != 5 || hotspots[1].Commits != 1 {
		t.Errorf("hotspot 1: got %+v, want {util.go, 5, 1}", hotspots[1])
	}
	// readme.md: 1+0 = 1 line, 1 commit
	if hotspots[2].Path != "readme.md" || hotspots[2].LinesChanged != 1 || hotspots[2].Commits != 1 {
		t.Errorf("hotspot 2: got %+v, want {readme.md, 1, 1}", hotspots[2])
	}
}

func TestFileHotspots_DateFiltering(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "jan commit")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC), "mar commit")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "bbb1", "main.go", 20, 10)

	// Only query January — should exclude the March commit.
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.FileHotspots(db, from, to)
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d: %v", len(hotspots), hotspots)
	}
	if hotspots[0].LinesChanged != 15 || hotspots[0].Commits != 1 {
		t.Errorf("got %+v, want {main.go, 15, 1}", hotspots[0])
	}
}

func TestFileHotspots_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.FileHotspots(db, from, to)
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 0 {
		t.Errorf("expected 0 hotspots, got %d", len(hotspots))
	}
}
