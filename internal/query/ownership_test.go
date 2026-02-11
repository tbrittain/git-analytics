package query_test

import (
	"math"
	"testing"
	"time"

	"git-analytics/internal/query"
)

func TestFileOwnerships_Basic(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 16, 12, 0, 0, 0, time.UTC), "second")

	// Alice: 60+15=75 lines on main.go, Bob: 20+5=25 lines
	insertFileStat(t, db, "aaa1", "main.go", 60, 15)
	insertFileStat(t, db, "bbb1", "main.go", 20, 5)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	results, err := query.FileOwnerships(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileOwnerships: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Path != "main.go" {
		t.Errorf("path: got %s, want main.go", r.Path)
	}
	if r.TopAuthorEmail != "alice@example.com" {
		t.Errorf("top author: got %s, want alice@example.com", r.TopAuthorEmail)
	}
	if math.Abs(r.TopAuthorPct-75.0) > 0.1 {
		t.Errorf("top author pct: got %.1f, want 75.0", r.TopAuthorPct)
	}
	if r.SecondAuthorEmail != "bob@example.com" {
		t.Errorf("second author: got %s, want bob@example.com", r.SecondAuthorEmail)
	}
	if math.Abs(r.SecondAuthorPct-25.0) > 0.1 {
		t.Errorf("second author pct: got %.1f, want 25.0", r.SecondAuthorPct)
	}
	if r.ContributorCount != 2 {
		t.Errorf("contributor count: got %d, want 2", r.ContributorCount)
	}
	if r.TotalLines != 100 {
		t.Errorf("total lines: got %d, want 100", r.TotalLines)
	}
}

func TestFileOwnerships_SingleContributor(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "solo")

	insertFileStat(t, db, "aaa1", "main.go", 50, 10)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	results, err := query.FileOwnerships(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileOwnerships: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if math.Abs(r.TopAuthorPct-100.0) > 0.1 {
		t.Errorf("top author pct: got %.1f, want 100.0", r.TopAuthorPct)
	}
	if r.SecondAuthorName != "" || r.SecondAuthorEmail != "" {
		t.Errorf("second author should be empty, got name=%q email=%q", r.SecondAuthorName, r.SecondAuthorEmail)
	}
	if r.SecondAuthorPct != 0 {
		t.Errorf("second author pct: got %.1f, want 0", r.SecondAuthorPct)
	}
	if r.ContributorCount != 1 {
		t.Errorf("contributor count: got %d, want 1", r.ContributorCount)
	}
}

func TestFileOwnerships_DateFiltering(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "jan commit")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC), "mar commit")

	insertFileStat(t, db, "aaa1", "main.go", 30, 10)
	insertFileStat(t, db, "bbb1", "main.go", 50, 20)

	// Only January — should only see Alice.
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	results, err := query.FileOwnerships(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileOwnerships: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.TopAuthorEmail != "alice@example.com" {
		t.Errorf("top author: got %s, want alice@example.com", r.TopAuthorEmail)
	}
	if math.Abs(r.TopAuthorPct-100.0) > 0.1 {
		t.Errorf("top author pct: got %.1f, want 100.0", r.TopAuthorPct)
	}
	if r.ContributorCount != 1 {
		t.Errorf("contributor count: got %d, want 1", r.ContributorCount)
	}
}

func TestFileOwnerships_ExcludeGlobs(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "aaa1", "generated.pb.go", 200, 0)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	results, err := query.FileOwnerships(db, from, to, []string{"*.pb.go"})
	if err != nil {
		t.Fatalf("FileOwnerships: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "main.go" {
		t.Errorf("expected main.go, got %s", results[0].Path)
	}
}

func TestFileOwnerships_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	results, err := query.FileOwnerships(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileOwnerships: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestFileOwnerships_MultipleFiles(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 16, 12, 0, 0, 0, time.UTC), "second")

	// main.go: Alice 90%, Bob 10% → top_author_pct = 90
	insertFileStat(t, db, "aaa1", "main.go", 90, 0)
	insertFileStat(t, db, "bbb1", "main.go", 10, 0)

	// util.go: Alice 50%, Bob 50% → top_author_pct = 50
	insertFileStat(t, db, "aaa1", "util.go", 25, 25)
	insertFileStat(t, db, "bbb1", "util.go", 25, 25)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	results, err := query.FileOwnerships(db, from, to, nil)
	if err != nil {
		t.Fatalf("FileOwnerships: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Should be sorted by top_author_pct descending: main.go (90%) first.
	if results[0].Path != "main.go" {
		t.Errorf("first result: got %s, want main.go", results[0].Path)
	}
	if math.Abs(results[0].TopAuthorPct-90.0) > 0.1 {
		t.Errorf("main.go top pct: got %.1f, want 90.0", results[0].TopAuthorPct)
	}
	if results[1].Path != "util.go" {
		t.Errorf("second result: got %s, want util.go", results[1].Path)
	}
	if math.Abs(results[1].TopAuthorPct-50.0) > 0.1 {
		t.Errorf("util.go top pct: got %.1f, want 50.0", results[1].TopAuthorPct)
	}
}
