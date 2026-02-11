package query_test

import (
	"testing"
	"time"

	"git-analytics/internal/query"
)

func TestGetDashboardStats_Basic(t *testing.T) {
	db := setupDB(t)

	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "first")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 16, 9, 0, 0, 0, time.UTC), "second")

	insertFileStat(t, db, "aaa1", "main.go", 10, 5)
	insertFileStat(t, db, "aaa1", "util.go", 3, 2)
	insertFileStat(t, db, "bbb1", "main.go", 20, 10)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	s, err := query.GetDashboardStats(db, from, to)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}

	if s.Commits != 2 {
		t.Errorf("Commits: got %d, want 2", s.Commits)
	}
	if s.Contributors != 2 {
		t.Errorf("Contributors: got %d, want 2", s.Contributors)
	}
	if s.Additions != 33 {
		t.Errorf("Additions: got %d, want 33", s.Additions)
	}
	if s.Deletions != 17 {
		t.Errorf("Deletions: got %d, want 17", s.Deletions)
	}
	if s.FilesChanged != 2 {
		t.Errorf("FilesChanged: got %d, want 2", s.FilesChanged)
	}
}

func TestGetDashboardStats_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	s, err := query.GetDashboardStats(db, from, to)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}

	if s.Commits != 0 || s.Contributors != 0 || s.Additions != 0 || s.Deletions != 0 || s.FilesChanged != 0 {
		t.Errorf("expected all zeros, got %+v", s)
	}
}

func TestGetDashboardStats_DateFiltering(t *testing.T) {
	db := setupDB(t)

	// Inside range
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "in range")
	insertFileStat(t, db, "aaa1", "main.go", 10, 5)

	// Outside range
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC), "out of range")
	insertFileStat(t, db, "bbb1", "other.go", 100, 50)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	s, err := query.GetDashboardStats(db, from, to)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}

	if s.Commits != 1 {
		t.Errorf("Commits: got %d, want 1", s.Commits)
	}
	if s.Additions != 10 {
		t.Errorf("Additions: got %d, want 10", s.Additions)
	}
}

func TestCommitsByHour_Basic(t *testing.T) {
	db := setupDB(t)

	// Two commits at hour 10, one at hour 14
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), "morning 1")
	insertCommit(t, db, "aaa2", "Alice", "alice@example.com",
		time.Date(2025, 1, 16, 10, 30, 0, 0, time.UTC), "morning 2")
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC), "afternoon")

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	buckets, err := query.CommitsByHour(db, from, to)
	if err != nil {
		t.Fatalf("CommitsByHour: %v", err)
	}

	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d: %v", len(buckets), buckets)
	}
	if buckets[0].Hour != 10 || buckets[0].Count != 2 {
		t.Errorf("bucket 0: got %+v, want {Hour:10 Count:2}", buckets[0])
	}
	if buckets[1].Hour != 14 || buckets[1].Count != 1 {
		t.Errorf("bucket 1: got %+v, want {Hour:14 Count:1}", buckets[1])
	}
}

func TestCommitsByHour_Empty(t *testing.T) {
	db := setupDB(t)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	buckets, err := query.CommitsByHour(db, from, to)
	if err != nil {
		t.Fatalf("CommitsByHour: %v", err)
	}

	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestCommitsByHour_DateFiltering(t *testing.T) {
	db := setupDB(t)

	// Inside range
	insertCommit(t, db, "aaa1", "Alice", "alice@example.com",
		time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC), "in range")

	// Outside range
	insertCommit(t, db, "bbb1", "Bob", "bob@example.com",
		time.Date(2025, 3, 1, 9, 0, 0, 0, time.UTC), "out of range")

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	buckets, err := query.CommitsByHour(db, from, to)
	if err != nil {
		t.Fatalf("CommitsByHour: %v", err)
	}

	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d: %v", len(buckets), buckets)
	}
	if buckets[0].Hour != 9 || buckets[0].Count != 1 {
		t.Errorf("bucket 0: got %+v, want {Hour:9 Count:1}", buckets[0])
	}
}
