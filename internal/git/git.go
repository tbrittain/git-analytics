package git

import "time"

// Commit holds the extracted analytics data for a single commit.
type Commit struct {
	Hash         string
	AuthorName   string
	AuthorEmail  string
	Date         time.Time
	Message      string
	FilesChanged []FileStat
}

// FileStat holds per-file change metrics for a commit.
type FileStat struct {
	Path      string
	Additions int
	Deletions int
}

// CommitIter yields commits one at a time. Callers must call Close when done.
type CommitIter interface {
	// Next returns the next commit, or nil, nil when exhausted.
	Next() (*Commit, error)
	Close()
}

// Repository is a data source for extracting commit analytics from a git repo.
type Repository interface {
	// Log returns an iterator over commits in reverse chronological order.
	// If sinceHash is non-empty, only commits after that hash are returned.
	Log(sinceHash string) (CommitIter, error)
	// HeadHash returns the current HEAD commit hash.
	HeadHash() (string, error)
	Close() error
}
