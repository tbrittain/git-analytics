package query_test

import (
	"database/sql"
	"math"
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

// --- TemporalHotspots tests ---

func TestTemporalHotspots_Basic(t *testing.T) {
	db := setupDB(t)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// Recent commit (1 day ago)
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC), "recent")
	// Old commit (30 days ago)
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC), "old")

	// Same churn for both files
	insertFileStat(t, db, "aaa1", "recent.go", 50, 50)
	insertFileStat(t, db, "bbb1", "old.go", 50, 50)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	hotspots, err := query.TemporalHotspots(db, from, to, 90, nil)
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 2 {
		t.Fatalf("expected 2 hotspots, got %d", len(hotspots))
	}

	// Recent file should score higher
	if hotspots[0].Path != "recent.go" {
		t.Errorf("expected recent.go first, got %s", hotspots[0].Path)
	}
	if hotspots[0].Score <= hotspots[1].Score {
		t.Errorf("expected first score > second: %f vs %f", hotspots[0].Score, hotspots[1].Score)
	}
}

func TestTemporalHotspots_DecayMath(t *testing.T) {
	db := setupDB(t)
	halfLife := 30.0
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// Commit exactly halfLifeDays before `to`
	commitTime := to.AddDate(0, 0, -int(halfLife))
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com", commitTime, "half")
	insertFileStat(t, db, "aaa1", "main.go", 50, 50) // 100 lines changed

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	hotspots, err := query.TemporalHotspots(db, from, to, halfLife, nil)
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d", len(hotspots))
	}

	// Score should be approximately linesChanged * 0.5
	expected := 100.0 * 0.5
	if math.Abs(hotspots[0].Score-expected) > 1.0 {
		t.Errorf("expected score ≈ %.1f, got %.4f", expected, hotspots[0].Score)
	}
}

func TestTemporalHotspots_SameDay(t *testing.T) {
	db := setupDB(t)
	to := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)

	// Commit on same day as `to`
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC), "today")
	insertFileStat(t, db, "aaa1", "main.go", 30, 20) // 50 lines

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	hotspots, err := query.TemporalHotspots(db, from, to, 90, nil)
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d", len(hotspots))
	}

	// Score should be very close to linesChanged (≈50)
	if math.Abs(hotspots[0].Score-50.0) > 1.0 {
		t.Errorf("expected score ≈ 50, got %.4f", hotspots[0].Score)
	}
}

func TestTemporalHotspots_DateFiltering(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "jan commit")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC), "mar commit")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "bbb1", "main.go", 20, 10)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.TemporalHotspots(db, from, to, 90, nil)
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d: %v", len(hotspots), hotspots)
	}
	if hotspots[0].LinesChanged != 15 {
		t.Errorf("expected 15 lines changed, got %d", hotspots[0].LinesChanged)
	}
}

func TestTemporalHotspots_ExcludeGlobs(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "aaa1", "generated.pb.go", 200, 0)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.TemporalHotspots(db, from, to, 90, []string{"*.pb.go"})
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d: %v", len(hotspots), hotspots)
	}
	if hotspots[0].Path != "main.go" {
		t.Errorf("expected main.go, got %s", hotspots[0].Path)
	}
}

func TestTemporalHotspots_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	hotspots, err := query.TemporalHotspots(db, from, to, 90, nil)
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 0 {
		t.Errorf("expected 0 hotspots, got %d", len(hotspots))
	}
}

func TestTemporalHotspots_ReordersByScore(t *testing.T) {
	db := setupDB(t)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// High churn, old file
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC), "old big change")
	insertFileStat(t, db, "aaa1", "big_old.go", 500, 500) // 1000 lines, 30 days ago

	// Low churn, recent file
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC), "recent small change")
	insertFileStat(t, db, "bbb1", "small_new.go", 10, 10) // 20 lines, 1 day ago

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Use a very short half-life so old changes decay heavily
	hotspots, err := query.TemporalHotspots(db, from, to, 5, nil)
	if err != nil {
		t.Fatalf("TemporalHotspots: %v", err)
	}

	if len(hotspots) != 2 {
		t.Fatalf("expected 2 hotspots, got %d", len(hotspots))
	}

	// With 5-day half-life, the 30-day-old file decays by ~2^6=64x,
	// so 1000/64 ≈ 15.6, while the 1-day-old file barely decays (20 * ~0.87 ≈ 17.4).
	// Recent small file should rank higher.
	if hotspots[0].Path != "small_new.go" {
		t.Errorf("expected small_new.go first (recency wins), got %s (scores: %f, %f)",
			hotspots[0].Path, hotspots[0].Score, hotspots[1].Score)
	}
}
