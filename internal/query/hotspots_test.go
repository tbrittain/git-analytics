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

	hotspots, err := query.FileHotspots(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 3 {
		t.Fatalf("expected 3 hotspots, got %d: %v", len(hotspots), hotspots)
	}

	// main.go: adds 10+20=30, dels 5+10=15, lines 45, 2 commits
	h := hotspots[0]
	if h.Path != "main.go" || h.LinesChanged != 45 || h.Additions != 30 || h.Deletions != 15 || h.Commits != 2 {
		t.Errorf("hotspot 0: got %+v, want {main.go, 45, 30, 15, 2}", h)
	}
	// util.go: adds 3, dels 2, lines 5, 1 commit
	h = hotspots[1]
	if h.Path != "util.go" || h.LinesChanged != 5 || h.Additions != 3 || h.Deletions != 2 || h.Commits != 1 {
		t.Errorf("hotspot 1: got %+v, want {util.go, 5, 3, 2, 1}", h)
	}
	// readme.md: adds 1, dels 0, lines 1, 1 commit
	h = hotspots[2]
	if h.Path != "readme.md" || h.LinesChanged != 1 || h.Additions != 1 || h.Deletions != 0 || h.Commits != 1 {
		t.Errorf("hotspot 2: got %+v, want {readme.md, 1, 1, 0, 1}", h)
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

	hotspots, err := query.FileHotspots(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d: %v", len(hotspots), hotspots)
	}
	if hotspots[0].LinesChanged != 15 || hotspots[0].Additions != 10 || hotspots[0].Deletions != 5 || hotspots[0].Commits != 1 {
		t.Errorf("got %+v, want {main.go, 15, 10, 5, 1}", hotspots[0])
	}
}

func TestFileHotspots_ExcludeGlobs(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "aaa1", "generated.pb.go", 200, 0)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.FileHotspots(db, from, to, []string{"*.pb.go"})
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d: %v", len(hotspots), hotspots)
	}
	if hotspots[0].Path != "main.go" {
		t.Errorf("expected main.go, got %s", hotspots[0].Path)
	}
}

func TestFileHotspots_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.FileHotspots(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileHotspots: %v", err)
	}

	if len(hotspots) != 0 {
		t.Errorf("expected 0 hotspots, got %d", len(hotspots))
	}
}
