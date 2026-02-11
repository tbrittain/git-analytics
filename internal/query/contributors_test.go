package query_test

import (
	"testing"
	"time"

	"git-analytics/internal/query"
)

func TestContributors_Basic(t *testing.T) {
	db := setupDB(t)

	// Alice: 2 commits, Bob: 1 commit
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "aaa2", "Alice", "alice@example.com",
		time.Date(2025, 1, 16, 12, 0, 0, 0, time.UTC), "second")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 17, 9, 0, 0, 0, time.UTC), "third")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "aaa1", "util.go", 3, 2)
	insertFileStat(t, db, "aaa2", "main.go", 20, 10)
	insertFileStat(t, db, "bbb1", "main.go", 8, 4)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	contributors, err := query.Contributors(db, from, to)
	if err != nil {
		t.Fatalf("Contributors: %v", err)
	}

	if len(contributors) != 2 {
		t.Fatalf("expected 2 contributors, got %d: %v", len(contributors), contributors)
	}

	// Alice first (2 commits), adds 10+3+20=33, dels 5+2+10=17
	c := contributors[0]
	if c.AuthorEmail != "alice@example.com" || c.Commits != 2 || c.Additions != 33 || c.Deletions != 17 {
		t.Errorf("contributor 0: got %+v, want {alice@example.com, Alice, 2, 33, 17}", c)
	}
	// Bob second (1 commit), adds 8, dels 4
	c = contributors[1]
	if c.AuthorEmail != "bob@example.com" || c.Commits != 1 || c.Additions != 8 || c.Deletions != 4 {
		t.Errorf("contributor 1: got %+v, want {bob@example.com, Bob, 1, 8, 4}", c)
	}
}

func TestContributors_CommitWithoutFileStats(t *testing.T) {
	db := setupDB(t)

	// A commit with no file_stats (e.g. merge commit)
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "merge commit")

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	contributors, err := query.Contributors(db, from, to)
	if err != nil {
		t.Fatalf("Contributors: %v", err)
	}

	if len(contributors) != 1 {
		t.Fatalf("expected 1 contributor, got %d: %v", len(contributors), contributors)
	}

	c := contributors[0]
	if c.Commits != 1 || c.Additions != 0 || c.Deletions != 0 {
		t.Errorf("got %+v, want {commits:1, additions:0, deletions:0}", c)
	}
}

func TestContributors_DateFiltering(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "jan commit")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC), "mar commit")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "bbb1", "main.go", 20, 10)

	// Only query January — should exclude Bob's March commit.
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	contributors, err := query.Contributors(db, from, to)
	if err != nil {
		t.Fatalf("Contributors: %v", err)
	}

	if len(contributors) != 1 {
		t.Fatalf("expected 1 contributor, got %d: %v", len(contributors), contributors)
	}
	if contributors[0].AuthorEmail != "alice@example.com" {
		t.Errorf("expected alice, got %s", contributors[0].AuthorEmail)
	}
}

func TestContributors_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	contributors, err := query.Contributors(db, from, to)
	if err != nil {
		t.Fatalf("Contributors: %v", err)
	}

	if len(contributors) != 0 {
		t.Errorf("expected 0 contributors, got %d", len(contributors))
	}
}
