package config

import (
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cfg.RecentRepos) != 0 {
		t.Fatalf("expected 0 recent repos, got %d", len(cfg.RecentRepos))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := &AppConfig{}
	cfg.AddRecent("/path/to/repo", "repo")
	cfg.AddRecent("/path/to/other", "other")

	if err := cfg.Save(dir); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(loaded.RecentRepos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(loaded.RecentRepos))
	}
	if loaded.RecentRepos[0].Name != "other" {
		t.Fatalf("expected first repo to be 'other', got %q", loaded.RecentRepos[0].Name)
	}
	if loaded.RecentRepos[1].Name != "repo" {
		t.Fatalf("expected second repo to be 'repo', got %q", loaded.RecentRepos[1].Name)
	}
}

func TestAddRecent_Dedup(t *testing.T) {
	cfg := &AppConfig{}
	cfg.AddRecent("/path/a", "a")
	cfg.AddRecent("/path/b", "b")
	cfg.AddRecent("/path/a", "a-updated")

	if len(cfg.RecentRepos) != 2 {
		t.Fatalf("expected 2 repos after dedup, got %d", len(cfg.RecentRepos))
	}
	if cfg.RecentRepos[0].Path != "/path/a" {
		t.Fatalf("expected first repo to be /path/a, got %q", cfg.RecentRepos[0].Path)
	}
	if cfg.RecentRepos[0].Name != "a-updated" {
		t.Fatalf("expected name to be updated to 'a-updated', got %q", cfg.RecentRepos[0].Name)
	}
}

func TestAddRecent_Limit(t *testing.T) {
	cfg := &AppConfig{}
	for i := 0; i < 12; i++ {
		cfg.AddRecent("/path/"+string(rune('a'+i)), string(rune('a'+i)))
	}

	if len(cfg.RecentRepos) != maxRecent {
		t.Fatalf("expected %d repos, got %d", maxRecent, len(cfg.RecentRepos))
	}
	// Most recently added should be first.
	if cfg.RecentRepos[0].Path != "/path/l" {
		t.Fatalf("expected first repo to be /path/l, got %q", cfg.RecentRepos[0].Path)
	}
}

func TestAddRecent_MRU(t *testing.T) {
	cfg := &AppConfig{}
	cfg.AddRecent("/path/a", "a")
	cfg.AddRecent("/path/b", "b")
	cfg.AddRecent("/path/c", "c")

	// Re-open "a" — should move to front.
	cfg.AddRecent("/path/a", "a")

	if cfg.RecentRepos[0].Path != "/path/a" {
		t.Fatalf("expected /path/a at front, got %q", cfg.RecentRepos[0].Path)
	}
	if cfg.RecentRepos[1].Path != "/path/c" {
		t.Fatalf("expected /path/c second, got %q", cfg.RecentRepos[1].Path)
	}
	if cfg.RecentRepos[2].Path != "/path/b" {
		t.Fatalf("expected /path/b third, got %q", cfg.RecentRepos[2].Path)
	}
}
