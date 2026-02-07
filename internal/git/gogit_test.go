package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"git-analytics/internal/git"
)

func TestOpenAndLog(t *testing.T) {
	repoPath := initTestRepo(t)

	repo, err := git.Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer repo.Close()

	headHash, err := repo.HeadHash()
	if err != nil {
		t.Fatalf("HeadHash: %v", err)
	}
	if headHash == "" {
		t.Fatal("HeadHash returned empty string")
	}

	iter, err := repo.Log("")
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	defer iter.Close()

	var commits []git.Commit
	for {
		c, err := iter.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if c == nil {
			break
		}
		commits = append(commits, *c)
	}

	// We created 2 commits in initTestRepo.
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	// Commits should be in reverse chronological order.
	second := commits[0]
	first := commits[1]

	if first.AuthorName != "Test User" {
		t.Errorf("expected author 'Test User', got %q", first.AuthorName)
	}
	if first.AuthorEmail != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", first.AuthorEmail)
	}
	if first.Message != "first commit\n" && first.Message != "first commit" {
		t.Errorf("unexpected message %q", first.Message)
	}
	if first.Date.IsZero() {
		t.Error("expected non-zero date")
	}

	// First commit should have 1 file added.
	if len(first.FilesChanged) != 1 {
		t.Fatalf("first commit: expected 1 file changed, got %d", len(first.FilesChanged))
	}
	if first.FilesChanged[0].Path != "hello.txt" {
		t.Errorf("expected file 'hello.txt', got %q", first.FilesChanged[0].Path)
	}
	if first.FilesChanged[0].Additions != 1 {
		t.Errorf("expected 1 addition, got %d", first.FilesChanged[0].Additions)
	}

	// Second commit should modify hello.txt.
	if second.Message != "second commit\n" && second.Message != "second commit" {
		t.Errorf("unexpected message %q", second.Message)
	}
	if len(second.FilesChanged) != 1 {
		t.Fatalf("second commit: expected 1 file changed, got %d", len(second.FilesChanged))
	}
}

func TestLogSinceHash(t *testing.T) {
	repoPath := initTestRepo(t)

	repo, err := git.Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer repo.Close()

	// Get all commits to find the first commit's hash.
	iter, err := repo.Log("")
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	var allCommits []git.Commit
	for {
		c, err := iter.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if c == nil {
			break
		}
		allCommits = append(allCommits, *c)
	}
	iter.Close()

	if len(allCommits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(allCommits))
	}

	// Log since the first commit — should only return the second commit.
	firstHash := allCommits[1].Hash // oldest commit
	iter2, err := repo.Log(firstHash)
	if err != nil {
		t.Fatalf("Log(sinceHash): %v", err)
	}
	defer iter2.Close()

	var newCommits []git.Commit
	for {
		c, err := iter2.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if c == nil {
			break
		}
		newCommits = append(newCommits, *c)
	}

	if len(newCommits) != 1 {
		t.Fatalf("expected 1 new commit, got %d", len(newCommits))
	}
	if newCommits[0].Hash != allCommits[0].Hash {
		t.Errorf("expected hash %s, got %s", allCommits[0].Hash, newCommits[0].Hash)
	}
}

func TestHeadHash(t *testing.T) {
	repoPath := initTestRepo(t)

	repo, err := git.Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer repo.Close()

	hash, err := repo.HeadHash()
	if err != nil {
		t.Fatalf("HeadHash: %v", err)
	}

	// SHA1 hex is 40 chars.
	if len(hash) != 40 {
		t.Errorf("expected 40-char hash, got %d chars: %q", len(hash), hash)
	}
}

// initTestRepo creates a temporary git repository with 2 commits for testing.
func initTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test User",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test User",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_AUTHOR_DATE="+time.Now().Add(-time.Hour).Format(time.RFC3339),
			"GIT_COMMITTER_DATE="+time.Now().Add(-time.Hour).Format(time.RFC3339),
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.name", "Test User")
	run("git", "config", "user.email", "test@example.com")

	// First commit.
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "hello.txt")
	run("git", "commit", "-m", "first commit")

	// Brief pause to ensure different commit timestamps.
	time.Sleep(time.Second)

	// Second commit.
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "hello.txt")
	run("git", "commit", "-m", "second commit")

	return dir
}
