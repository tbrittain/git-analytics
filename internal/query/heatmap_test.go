package query_test

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"git-analytics/internal/query"
	"git-analytics/internal/store"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(store.SchemaSQL); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func insertCommit(t *testing.T, db *sql.DB, hash, name, email string, at time.Time, msg string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO commits (hash, author_name, author_email, committed_at, message) VALUES (?, ?, ?, ?, ?)`,
		hash, name, email, at, msg,
	)
	if err != nil {
		t.Fatalf("insert commit: %v", err)
	}
}

func TestCommitHeatmap_Basic(t *testing.T) {
	db := setupDB(t)

	// Two commits on Jan 15, one on Jan 16.
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "aaa2", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC), "second")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 16, 9, 0, 0, 0, time.UTC), "third")

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	days, err := query.CommitHeatmap(db, from, to, "")
	if err != nil {
		t.Fatalf("CommitHeatmap: %v", err)
	}

	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d: %v", len(days), days)
	}
	if days[0].Date != "2025-01-15" || days[0].Count != 2 {
		t.Errorf("day 0: got %+v, want {2025-01-15, 2}", days[0])
	}
	if days[1].Date != "2025-01-16" || days[1].Count != 1 {
		t.Errorf("day 1: got %+v, want {2025-01-16, 1}", days[1])
	}
}

func TestCommitHeatmap_FilterByEmail(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "alice commit")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC), "bob commit")

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	days, err := query.CommitHeatmap(db, from, to, "alice@example.com")
	if err != nil {
		t.Fatalf("CommitHeatmap: %v", err)
	}

	if len(days) != 1 {
		t.Fatalf("expected 1 day, got %d: %v", len(days), days)
	}
	if days[0].Count != 1 {
		t.Errorf("expected count 1, got %d", days[0].Count)
	}
}

func TestCommitHeatmap_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	days, err := query.CommitHeatmap(db, from, to, "")
	if err != nil {
		t.Fatalf("CommitHeatmap: %v", err)
	}

	if len(days) != 0 {
		t.Errorf("expected 0 days, got %d", len(days))
	}
}
