package query_test

import (
	"testing"
	"time"

	"git-analytics/internal/query"
)

func TestCoChanges_Basic(t *testing.T) {
	db := setupDB(t)

	// 3 commits: A+B always together, C only in commit 3
	insertCommit(t, db, "c1", "Alice", "alice@example.com",
		time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "c2", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "second")
	insertCommit(t, db, "c3", "Alice", "alice@example.com",
		time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC), "third")

	insertFileStat(t, db, "c1", "a.go", 10, 5)
	insertFileStat(t, db, "c1", "b.go", 5, 2)

	insertFileStat(t, db, "c2", "a.go", 8, 3)
	insertFileStat(t, db, "c2", "b.go", 4, 1)

	insertFileStat(t, db, "c3", "a.go", 6, 2)
	insertFileStat(t, db, "c3", "b.go", 3, 1)
	insertFileStat(t, db, "c3", "c.go", 20, 10)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	pairs, err := query.CoChanges(db, from, to, 1, 100, nil)
	if err != nil {
		t.Fatalf("CoChanges: %v", err)
	}

	// Pairs: a↔b (3), a↔c (1), b↔c (1)
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d: %v", len(pairs), pairs)
	}

	// First pair should be a↔b with 3 co-changes
	p := pairs[0]
	if p.FileA != "a.go" || p.FileB != "b.go" {
		t.Errorf("pair 0: expected a.go↔b.go, got %s↔%s", p.FileA, p.FileB)
	}
	if p.CoChangeCount != 3 {
		t.Errorf("pair 0: expected co_change_count=3, got %d", p.CoChangeCount)
	}
	if p.CommitsA != 3 || p.CommitsB != 3 {
		t.Errorf("pair 0: expected commits 3/3, got %d/%d", p.CommitsA, p.CommitsB)
	}
	// coupling ratio = 3 / min(3,3) = 1.0
	if p.CouplingRatio != 1.0 {
		t.Errorf("pair 0: expected coupling_ratio=1.0, got %f", p.CouplingRatio)
	}
}

func TestCoChanges_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	pairs, err := query.CoChanges(db, from, to, 1, 100, nil)
	if err != nil {
		t.Fatalf("CoChanges: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestCoChanges_MinCount(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "c1", "Alice", "alice@example.com",
		time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "c2", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "second")

	// a+b in both commits, c only in commit 1
	insertFileStat(t, db, "c1", "a.go", 10, 5)
	insertFileStat(t, db, "c1", "b.go", 5, 2)
	insertFileStat(t, db, "c1", "c.go", 3, 1)

	insertFileStat(t, db, "c2", "a.go", 8, 3)
	insertFileStat(t, db, "c2", "b.go", 4, 1)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// minCount=2 should exclude pairs with c.go (only 1 co-change each)
	pairs, err := query.CoChanges(db, from, to, 2, 100, nil)
	if err != nil {
		t.Fatalf("CoChanges: %v", err)
	}

	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d: %v", len(pairs), pairs)
	}
	if pairs[0].FileA != "a.go" || pairs[0].FileB != "b.go" {
		t.Errorf("expected a.go↔b.go, got %s↔%s", pairs[0].FileA, pairs[0].FileB)
	}
}

func TestCoChanges_DateFiltering(t *testing.T) {
	db := setupDB(t)

	// Commit in January
	insertCommit(t, db, "c1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "jan")
	insertFileStat(t, db, "c1", "a.go", 10, 5)
	insertFileStat(t, db, "c1", "b.go", 5, 2)

	// Commit in March (outside range)
	insertCommit(t, db, "c2", "Alice", "alice@example.com",
		time.Date(2025, 3, 10, 10, 0, 0, 0, time.UTC), "mar")
	insertFileStat(t, db, "c2", "a.go", 8, 3)
	insertFileStat(t, db, "c2", "b.go", 4, 1)

	// Query only January
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	pairs, err := query.CoChanges(db, from, to, 1, 100, nil)
	if err != nil {
		t.Fatalf("CoChanges: %v", err)
	}

	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d: %v", len(pairs), pairs)
	}
	if pairs[0].CoChangeCount != 1 {
		t.Errorf("expected co_change_count=1, got %d", pairs[0].CoChangeCount)
	}
}

func TestCoChanges_ExcludeGlobs(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "c1", "Alice", "alice@example.com",
		time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "c2", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "second")

	insertFileStat(t, db, "c1", "main.go", 10, 5)
	insertFileStat(t, db, "c1", "generated.pb.go", 200, 0)
	insertFileStat(t, db, "c1", "util.go", 3, 1)

	insertFileStat(t, db, "c2", "main.go", 8, 3)
	insertFileStat(t, db, "c2", "generated.pb.go", 150, 0)
	insertFileStat(t, db, "c2", "util.go", 2, 1)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	// Exclude *.pb.go — should remove all pairs involving generated.pb.go
	pairs, err := query.CoChanges(db, from, to, 1, 100, []string{"*.pb.go"})
	if err != nil {
		t.Fatalf("CoChanges: %v", err)
	}

	// Only main.go↔util.go should remain
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d: %v", len(pairs), pairs)
	}
	if pairs[0].FileA != "main.go" || pairs[0].FileB != "util.go" {
		t.Errorf("expected main.go↔util.go, got %s↔%s", pairs[0].FileA, pairs[0].FileB)
	}
}

func TestCoChanges_Limit(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "c1", "Alice", "alice@example.com",
		time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC), "first")

	// 3 files in one commit → 3 pairs
	insertFileStat(t, db, "c1", "a.go", 10, 5)
	insertFileStat(t, db, "c1", "b.go", 5, 2)
	insertFileStat(t, db, "c1", "c.go", 3, 1)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	pairs, err := query.CoChanges(db, from, to, 1, 2, nil)
	if err != nil {
		t.Fatalf("CoChanges: %v", err)
	}

	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs (limit), got %d: %v", len(pairs), pairs)
	}
}
