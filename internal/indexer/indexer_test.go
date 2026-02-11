package indexer_test

import (
	"fmt"
	"testing"
	"time"

	"git-analytics/internal/git"
	"git-analytics/internal/indexer"
)

// fakeRepo implements git.Repository for testing.
type fakeRepo struct {
	headHash string
	commits  []git.Commit
}

func (r *fakeRepo) HeadHash() (string, error) {
	return r.headHash, nil
}

func (r *fakeRepo) Log(sinceHash string) (git.CommitIter, error) {
	var filtered []git.Commit
	for _, c := range r.commits {
		if sinceHash != "" && c.Hash == sinceHash {
			break
		}
		filtered = append(filtered, c)
	}
	return &fakeIter{commits: filtered}, nil
}

func (r *fakeRepo) RepoName() string      { return "fake-repo" }
func (r *fakeRepo) CurrentBranch() string { return "main" }
func (r *fakeRepo) Close() error          { return nil }

// fakeIter implements git.CommitIter for testing.
type fakeIter struct {
	commits []git.Commit
	pos     int
}

func (it *fakeIter) Next() (*git.Commit, error) {
	if it.pos >= len(it.commits) {
		return nil, nil
	}
	c := it.commits[it.pos]
	it.pos++
	return &c, nil
}

func (it *fakeIter) Close() {}

// fakeStore implements store.Store for testing.
type fakeStore struct {
	lastIndexed     string
	insertedBatches [][]git.Commit
	initCalled      bool
}

func (s *fakeStore) Init() error {
	s.initCalled = true
	return nil
}

func (s *fakeStore) InsertCommits(commits []git.Commit) error {
	batch := make([]git.Commit, len(commits))
	copy(batch, commits)
	s.insertedBatches = append(s.insertedBatches, batch)
	return nil
}

func (s *fakeStore) GetLastIndexedCommit() (string, error) {
	return s.lastIndexed, nil
}

func (s *fakeStore) SetLastIndexedCommit(hash string) error {
	s.lastIndexed = hash
	return nil
}

func (s *fakeStore) Close() error { return nil }

func TestIndexFullRepo(t *testing.T) {
	commits := makeCommits(3)
	repo := &fakeRepo{
		headHash: commits[0].Hash,
		commits:  commits,
	}
	store := &fakeStore{}

	idx := indexer.New(repo, store)
	if err := idx.Index(); err != nil {
		t.Fatalf("Index: %v", err)
	}

	// All 3 commits should be inserted.
	total := 0
	for _, batch := range store.insertedBatches {
		total += len(batch)
	}
	if total != 3 {
		t.Errorf("expected 3 commits inserted, got %d", total)
	}

	// Last indexed should be set to head.
	if store.lastIndexed != commits[0].Hash {
		t.Errorf("expected last indexed %q, got %q", commits[0].Hash, store.lastIndexed)
	}
}

func TestIndexIncremental(t *testing.T) {
	commits := makeCommits(3)
	repo := &fakeRepo{
		headHash: commits[0].Hash,
		commits:  commits,
	}

	// Pretend the oldest commit was already indexed.
	store := &fakeStore{lastIndexed: commits[2].Hash}

	idx := indexer.New(repo, store)
	if err := idx.Index(); err != nil {
		t.Fatalf("Index: %v", err)
	}

	// Only 2 new commits should be inserted.
	total := 0
	for _, batch := range store.insertedBatches {
		total += len(batch)
	}
	if total != 2 {
		t.Errorf("expected 2 commits inserted, got %d", total)
	}
}

func TestIndexNoOp(t *testing.T) {
	commits := makeCommits(1)
	repo := &fakeRepo{
		headHash: commits[0].Hash,
		commits:  commits,
	}

	// Already up to date.
	store := &fakeStore{lastIndexed: commits[0].Hash}

	idx := indexer.New(repo, store)
	if err := idx.Index(); err != nil {
		t.Fatalf("Index: %v", err)
	}

	// Nothing should be inserted.
	if len(store.insertedBatches) != 0 {
		t.Errorf("expected no inserts, got %d batches", len(store.insertedBatches))
	}
}

// makeCommits creates n fake commits in reverse chronological order.
func makeCommits(n int) []git.Commit {
	commits := make([]git.Commit, n)
	for i := 0; i < n; i++ {
		commits[i] = git.Commit{
			Hash:        fmt.Sprintf("%040d", n-i),
			AuthorName:  "Test",
			AuthorEmail: "test@example.com",
			Date:        time.Now().Add(-time.Duration(i) * time.Hour),
			Message:     fmt.Sprintf("commit %d", n-i),
			FilesChanged: []git.FileStat{
				{Path: "file.go", Additions: 10, Deletions: 5},
			},
		}
	}
	return commits
}
