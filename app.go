package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	_ "modernc.org/sqlite"

	"git-analytics/internal/git"
	"git-analytics/internal/indexer"
	"git-analytics/internal/query"
	"git-analytics/internal/store"
	sqlitestore "git-analytics/internal/store/sqlite"
)

// App struct
type App struct {
	ctx   context.Context
	repo  git.Repository
	store store.Store
	db    *sql.DB
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

// SelectDirectory opens a native OS folder picker and returns the selected path.
// Returns an empty string if the user cancels or an error occurs.
func (a *App) SelectDirectory() string {
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Git Repository",
	})
	if err != nil {
		return ""
	}
	return path
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.repo != nil {
		a.repo.Close()
	}
	if a.store != nil {
		a.store.Close()
	}
	if a.db != nil {
		a.db.Close()
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
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}

	repo, err := git.Open(path)
	if err != nil {
		return fmt.Errorf("opening repository: %w", err)
	}

	dbPath := filepath.Join(path, ".git-analytics.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		repo.Close()
		return fmt.Errorf("opening database: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		repo.Close()
		db.Close()
		return fmt.Errorf("setting WAL mode: %w", err)
	}

	s := sqlitestore.NewFromDB(db)
	if err := s.Init(); err != nil {
		repo.Close()
		db.Close()
		return fmt.Errorf("initializing schema: %w", err)
	}

	if err := addToGitExclude(path, ".git-analytics.db"); err != nil {
		repo.Close()
		db.Close()
		return fmt.Errorf("updating git exclude: %w", err)
	}

	a.repo = repo
	a.store = s
	a.db = db

	idx := indexer.New(repo, s)
	if err := idx.Index(); err != nil {
		return fmt.Errorf("indexing: %w", err)
	}

	return nil
}

// CommitHeatmap returns per-day commit counts between the given dates.
// Dates should be in "2006-01-02" format. An empty email returns counts for
// all authors.
func (a *App) CommitHeatmap(fromDate, toDate, email string) ([]query.HeatmapDay, error) {
	if a.db == nil {
		return nil, fmt.Errorf("no repository open")
	}

	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return nil, fmt.Errorf("parsing from date: %w", err)
	}
	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return nil, fmt.Errorf("parsing to date: %w", err)
	}

	return query.CommitHeatmap(a.db, from, to, email)
}

// FileHotspots returns per-file churn (lines changed) and commit counts
// between the given dates. Dates should be in "2006-01-02" format.
func (a *App) FileHotspots(fromDate, toDate string) ([]query.FileHotspot, error) {
	if a.db == nil {
		return nil, fmt.Errorf("no repository open")
	}

	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return nil, fmt.Errorf("parsing from date: %w", err)
	}
	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return nil, fmt.Errorf("parsing to date: %w", err)
	}

	return query.FileHotspots(a.db, from, to)
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
