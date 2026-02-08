package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"git-analytics/internal/git"
	"git-analytics/internal/indexer"
	"git-analytics/internal/store"
	sqlitestore "git-analytics/internal/store/sqlite"
)

// App struct
type App struct {
	ctx   context.Context
	repo  git.Repository
	store store.Store
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.repo != nil {
		a.repo.Close()
	}
	if a.store != nil {
		a.store.Close()
	}
}

// OpenRepository opens a git repository at the given path, initializes the
// analytics database, and runs the indexer.
func (a *App) OpenRepository(path string) error {
	// Close any previously opened resources.
	if a.repo != nil {
		a.repo.Close()
		a.repo = nil
	}
	if a.store != nil {
		a.store.Close()
		a.store = nil
	}

	repo, err := git.Open(path)
	if err != nil {
		return fmt.Errorf("opening repository: %w", err)
	}

	dbPath := filepath.Join(path, ".git-analytics.db")
	s, err := sqlitestore.Open(dbPath)
	if err != nil {
		repo.Close()
		return fmt.Errorf("opening database: %w", err)
	}

	if err := s.Init(); err != nil {
		repo.Close()
		s.Close()
		return fmt.Errorf("initializing schema: %w", err)
	}

	if err := addToGitExclude(path, ".git-analytics.db"); err != nil {
		repo.Close()
		s.Close()
		return fmt.Errorf("updating git exclude: %w", err)
	}

	a.repo = repo
	a.store = s

	idx := indexer.New(repo, s)
	if err := idx.Index(); err != nil {
		return fmt.Errorf("indexing: %w", err)
	}

	return nil
}

// addToGitExclude adds a pattern to .git/info/exclude if it's not already present.
func addToGitExclude(repoPath, pattern string) error {
	excludePath := filepath.Join(repoPath, ".git", "info", "exclude")

	existing, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if pattern is already in the file.
	lines := string(existing)
	for _, line := range splitLines(lines) {
		if line == pattern {
			return nil
		}
	}

	// Ensure we start on a new line.
	suffix := "\n"
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		suffix = "\n" + suffix
	}

	f, err := os.OpenFile(excludePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(suffix + pattern + "\n")
	return err
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
