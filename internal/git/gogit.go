package git

import (
	"errors"
	"io"
	"path/filepath"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/storer"
)

// goGitRepo implements Repository using go-git.
type goGitRepo struct {
	repo *gogit.Repository
	path string
}

// Open opens an existing git repository on disk.
func Open(path string) (Repository, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &goGitRepo{repo: repo, path: path}, nil
}

func (r *goGitRepo) RepoName() string {
	return filepath.Base(r.path)
}

func (r *goGitRepo) CurrentBranch() string {
	ref, err := r.repo.Head()
	if err != nil {
		return "HEAD"
	}
	return ref.Name().Short()
}

func (r *goGitRepo) HeadHash() (string, error) {
	ref, err := r.repo.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}

func (r *goGitRepo) Log(sinceHash string) (CommitIter, error) {
	opts := &gogit.LogOptions{
		Order: gogit.LogOrderCommitterTime,
	}

	if sinceHash != "" {
		hash, ok := plumbing.FromHex(sinceHash)
		if !ok {
			return nil, &InvalidHashError{Hash: sinceHash}
		}
		opts.To = hash
	}

	iter, err := r.repo.Log(opts)
	if err != nil {
		return nil, err
	}

	return &goGitCommitIter{
		iter:      iter,
		sinceHash: sinceHash,
	}, nil
}

func (r *goGitRepo) Close() error {
	return nil
}

// InvalidHashError is returned when a hash string cannot be parsed.
type InvalidHashError struct {
	Hash string
}

func (e *InvalidHashError) Error() string {
	return "invalid git hash: " + e.Hash
}

// goGitCommitIter implements CommitIter using go-git's commit iterator.
type goGitCommitIter struct {
	iter      object.CommitIter
	sinceHash string
}

func (it *goGitCommitIter) Next() (*Commit, error) {
	for {
		c, err := it.iter.Next()
		if err != nil {
			if err == io.EOF {
				return nil, nil
			}
			// storer.ErrStop is returned when the To/TailHash commit is
			// reached. That commit is the one we already indexed, so skip it.
			if errors.Is(err, storer.ErrStop) {
				return nil, nil
			}
			return nil, err
		}
		if c == nil {
			return nil, nil
		}

		stats, err := c.Stats()
		if err != nil {
			return nil, err
		}

		files := make([]FileStat, len(stats))
		for i, s := range stats {
			files[i] = FileStat{
				Path:      s.Name,
				Additions: s.Addition,
				Deletions: s.Deletion,
			}
		}

		return &Commit{
			Hash:         c.Hash.String(),
			AuthorName:   c.Author.Name,
			AuthorEmail:  c.Author.Email,
			Date:         c.Author.When,
			Message:      c.Message,
			FilesChanged: files,
		}, nil
	}
}

func (it *goGitCommitIter) Close() {
	it.iter.Close()
}
