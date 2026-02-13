package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	appName    = "git-analytics"
	configFile = "config.json"
	maxRecent  = 10
)

// RecentRepo records a previously opened repository.
type RecentRepo struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	OpenedAt time.Time `json:"opened_at"`
}

// AppConfig holds persistent application settings.
type AppConfig struct {
	RecentRepos []RecentRepo `json:"recent_repos"`
}

// DefaultConfigDir returns the platform-specific config directory for the app.
func DefaultConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// Load reads the config file from configDir. If the file does not exist,
// an empty config is returned without error.
func Load(configDir string) (*AppConfig, error) {
	path := filepath.Join(configDir, configFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &AppConfig{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &AppConfig{}, nil
	}
	return &cfg, nil
}

// Save writes the config to configDir, creating the directory if needed.
func (c *AppConfig) Save(configDir string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, configFile), data, 0644)
}

// AddRecent adds or updates a repo in the recent list. If the path already
// exists, its timestamp is updated and it moves to the front. The list is
// trimmed to the 10 most recent entries.
func (c *AppConfig) AddRecent(path, name string) {
	now := time.Now()

	// Remove existing entry with the same path.
	filtered := make([]RecentRepo, 0, len(c.RecentRepos))
	for _, r := range c.RecentRepos {
		if r.Path != path {
			filtered = append(filtered, r)
		}
	}

	// Prepend the new/updated entry.
	c.RecentRepos = append([]RecentRepo{{
		Path:     path,
		Name:     name,
		OpenedAt: now,
	}}, filtered...)

	// Trim to limit.
	if len(c.RecentRepos) > maxRecent {
		c.RecentRepos = c.RecentRepos[:maxRecent]
	}
}

// RemoveRecent removes the repo with the given path from the recent list.
func (c *AppConfig) RemoveRecent(path string) {
	filtered := make([]RecentRepo, 0, len(c.RecentRepos))
	for _, r := range c.RecentRepos {
		if r.Path != path {
			filtered = append(filtered, r)
		}
	}
	c.RecentRepos = filtered
}
