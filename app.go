package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	_ "modernc.org/sqlite"

	"git-analytics/internal/config"
	"git-analytics/internal/git"
	"git-analytics/internal/indexer"
	"git-analytics/internal/query"
	"git-analytics/internal/store"
	sqlitestore "git-analytics/internal/store/sqlite"
)

// App struct
type App struct {
	ctx       context.Context
	repo      git.Repository
	store     store.Store
	db        *sql.DB
	configDir string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if dir, err := config.DefaultConfigDir(); err == nil {
		a.configDir = dir
	}
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

	// Persist this repo in the recent list.
	if a.configDir != "" {
		cfg, _ := config.Load(a.configDir)
		cfg.AddRecent(path, repo.RepoName())
		_ = cfg.Save(a.configDir)
	}

	return nil
}

// RecentRepos returns the list of recently opened repositories.
func (a *App) RecentRepos() ([]config.RecentRepo, error) {
	if a.configDir == "" {
		return nil, fmt.Errorf("config directory unavailable")
	}
	cfg, err := config.Load(a.configDir)
	if err != nil {
		return nil, err
	}
	// Filter out repos whose paths no longer exist on disk.
	valid := make([]config.RecentRepo, 0, len(cfg.RecentRepos))
	for _, r := range cfg.RecentRepos {
		if _, err := os.Stat(r.Path); err == nil {
			valid = append(valid, r)
		}
	}
	return valid, nil
}

// RemoveRecentRepo removes a repository from the recent list.
func (a *App) RemoveRecentRepo(path string) error {
	if a.configDir == "" {
		return fmt.Errorf("config directory unavailable")
	}
	cfg, err := config.Load(a.configDir)
	if err != nil {
		return err
	}
	cfg.RemoveRecent(path)
	return cfg.Save(a.configDir)
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
// Files matching any of the excludeGlobs patterns are omitted.
func (a *App) FileHotspots(fromDate, toDate string, excludeGlobs []string) ([]query.FileHotspot, error) {
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

	return query.FileHotspots(a.db, from, to, excludeGlobs)
}

// Contributors returns per-author commit counts, additions, and deletions
// between the given dates. Dates should be in "2006-01-02" format.
// Files matching any of the excludeGlobs patterns are excluded from stats.
func (a *App) Contributors(fromDate, toDate string, excludeGlobs []string) ([]query.Contributor, error) {
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

	return query.Contributors(a.db, from, to, excludeGlobs)
}

// FileOwnerships returns per-file ownership analysis showing the dominant
// contributors between the given dates. Dates should be in "2006-01-02" format.
// Files matching any of the excludeGlobs patterns are omitted.
func (a *App) FileOwnerships(fromDate, toDate string, excludeGlobs []string) ([]query.FileOwnership, error) {
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

	return query.FileOwnerships(a.db, from, to, excludeGlobs)
}

// TemporalHotspots returns per-file churn weighted by recency (exponential
// decay) between the given dates. Dates should be in "2006-01-02" format.
// halfLifeDays controls how fast old changes decay. Files matching any of
// the excludeGlobs patterns are omitted.
func (a *App) TemporalHotspots(fromDate, toDate string, halfLifeDays float64, excludeGlobs []string) ([]query.TemporalHotspot, error) {
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

	return query.TemporalHotspots(a.db, from, to, halfLifeDays, excludeGlobs)
}

// CoChanges returns file pairs that frequently change together in commits
// between the given dates. Dates should be in "2006-01-02" format.
// Only pairs with at least minCount shared commits are returned, up to limit.
// Files matching any of the excludeGlobs patterns are omitted.
func (a *App) CoChanges(fromDate, toDate string, minCount int, limit int, excludeGlobs []string) ([]query.CoChangePair, error) {
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

	return query.CoChanges(a.db, from, to, minCount, limit, excludeGlobs)
}

// RepoInfo holds metadata about the currently opened repository.
type RepoInfo struct {
	Name          string `json:"name"`
	Branch        string `json:"branch"`
	HeadHash      string `json:"head_hash"`
	LastAuthor    string `json:"last_author"`
	LastEmail     string `json:"last_email"`
	LastMessage   string `json:"last_message"`
	LastCommitAge string `json:"last_commit_age"`
}

// RepoInfo returns metadata about the currently opened repository.
func (a *App) RepoInfo() (*RepoInfo, error) {
	if a.repo == nil || a.db == nil {
		return nil, fmt.Errorf("no repository open")
	}

	hash, err := a.repo.HeadHash()
	if err != nil {
		return nil, fmt.Errorf("reading HEAD: %w", err)
	}

	info := &RepoInfo{
		Name:     a.repo.RepoName(),
		Branch:   a.repo.CurrentBranch(),
		HeadHash: hash[:min(7, len(hash))],
	}

	var authorName, authorEmail, message, committedAt string
	err = a.db.QueryRow(
		`SELECT author_name, author_email, message, committed_at
		 FROM commits ORDER BY committed_at DESC LIMIT 1`,
	).Scan(&authorName, &authorEmail, &message, &committedAt)
	if err == sql.ErrNoRows {
		return info, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying last commit: %w", err)
	}

	info.LastAuthor = authorName
	info.LastEmail = authorEmail
	info.LastMessage = strings.TrimSpace(message)

	t, err := time.Parse(time.RFC3339, committedAt)
	if err != nil {
		trimmed := committedAt
		if idx := strings.LastIndex(trimmed, " "); idx > 0 {
			trimmed = trimmed[:idx]
		}
		t, err = time.Parse("2006-01-02 15:04:05 -0700", trimmed)
		if err != nil {
			info.LastCommitAge = committedAt
			return info, nil
		}
	}
	info.LastCommitAge = relativeTime(t)

	return info, nil
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// DashboardStats returns aggregate commit and file-change stats between the
// given dates. Dates should be in "2006-01-02" format.
func (a *App) DashboardStats(fromDate, toDate string) (*query.DashboardStats, error) {
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

	return query.GetDashboardStats(a.db, from, to)
}

// CommitsByHour returns per-hour commit counts between the given dates.
// Dates should be in "2006-01-02" format.
func (a *App) CommitsByHour(fromDate, toDate string) ([]query.HourBucket, error) {
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

	return query.CommitsByHour(a.db, from, to)
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
