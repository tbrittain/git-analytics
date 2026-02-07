package indexer

import (
	"git-analytics/internal/git"
	"git-analytics/internal/store"
)

const batchSize = 500

// Indexer is the data pipeline that reads commits from a git repository
// and persists them into a store.
type Indexer struct {
	repo  git.Repository
	store store.Store
}

// New creates a new Indexer.
func New(repo git.Repository, store store.Store) *Indexer {
	return &Indexer{repo: repo, store: store}
}

// Index reads all new commits from the repository and writes them to the store.
// It resumes from the last indexed commit if one exists.
func (idx *Indexer) Index() error {
	sinceHash, err := idx.store.GetLastIndexedCommit()
	if err != nil {
		return err
	}

	headHash, err := idx.repo.HeadHash()
	if err != nil {
		return err
	}

	// Nothing to do if HEAD hasn't changed.
	if sinceHash == headHash {
		return nil
	}

	iter, err := idx.repo.Log(sinceHash)
	if err != nil {
		return err
	}
	defer iter.Close()

	batch := make([]git.Commit, 0, batchSize)

	for {
		commit, err := iter.Next()
		if err != nil {
			return err
		}
		if commit == nil {
			break
		}

		batch = append(batch, *commit)

		if len(batch) >= batchSize {
			if err := idx.store.InsertCommits(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// Flush remaining commits.
	if len(batch) > 0 {
		if err := idx.store.InsertCommits(batch); err != nil {
			return err
		}
	}

	return idx.store.SetLastIndexedCommit(headHash)
}
