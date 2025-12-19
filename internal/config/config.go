package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config represents the application settings.
type Config struct {
	Exclusions          []string `json:"exclusions"`            // List of excluded directories/files
	Masks               []string `json:"masks"`                 // File glob masks (e.g., *.tmp)
	ShowPassed          bool     `json:"show_passed"`           // Whether to display successful tests
	AutoWatch           bool     `json:"auto_watch"`            // Enable automatic test re-run on changes
	ShowNotifications   bool     `json:"show_notifications"`    // Toggle system notifications
	NotifyOnlyOnFailure bool     `json:"notify_only_on_failure"` // Only notify when tests fail
	AlwaysOnTop         bool     `json:"always_on_top"`         // Keep window on top of others
	AutoCopyErrors      bool     `json:"auto_copy_errors"`      // Automatically copy errors to clipboard
	Lang                string   `json:"lang"`                  // UI language ("ru" or "en")
}

var (
	instance *Config
	mu       sync.RWMutex
	cfgPath  = "go_test_runner.cfg"
)

// Load retrieves the configuration from file or returns defaults.
func Load() *Config {
	mu.RLock()
	if instance != nil {
		defer mu.RUnlock()
		return instance
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return instance
	}

	instance = &Config{
		Exclusions:          []string{".git", "node_modules", "vendor", "frontend", "build"},
		ShowPassed:          true,
		AutoWatch:           true,
		ShowNotifications:   true,
		NotifyOnlyOnFailure: false,
		AlwaysOnTop:         false,
		AutoCopyErrors:      false,
		Lang:                "ru",
	}

	data, err := os.ReadFile(cfgPath)
	if err == nil {
		_ = json.Unmarshal(data, instance)
	}

	if instance.Lang == "" {
		instance.Lang = "ru"
	}

	return instance
}

// Save persists the provided configuration to disk.
func Save(newCfg *Config) error {
	mu.Lock()
	defer mu.Unlock()

	instance = newCfg
	data, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0644)
}

// IsExcluded checks if a given path should be ignored based on exclusions and masks.
func (c *Config) IsExcluded(path string, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return false
	}

	cleanPath := filepath.ToSlash(rel)
	lowerPath := strings.ToLower(cleanPath)
	segments := strings.Split(lowerPath, "/")

	for _, ex := range c.Exclusions {
		lowEx := strings.ToLower(ex)
		for _, s := range segments {
			if s == lowEx {
				return true
			}
		}
	}

	base := strings.ToLower(filepath.Base(cleanPath))
	for _, mask := range c.Masks {
		match, _ := filepath.Match(strings.ToLower(mask), base)
		if match {
			return true
		}
	}

	return false
}
