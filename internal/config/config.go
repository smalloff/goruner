package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	Exclusions        []string `json:"exclusions"` // Folders/Files
	Masks             []string `json:"masks"`      // *.tmp, etc
	ShowPassed        bool     `json:"show_passed"`
	AutoWatch         bool     `json:"auto_watch"`
	ShowNotifications bool     `json:"show_notifications"`
	NotifyOnlyOnFailure bool    `json:"notify_only_on_failure"`
	AlwaysOnTop         bool    `json:"always_on_top"`
	AutoCopyErrors      bool    `json:"auto_copy_errors"`
	Lang              string   `json:"lang"`       // "ru" or "en"
}

var (
	cfg     *Config
	mu      sync.RWMutex
	cfgPath = "go_test_runner.cfg"
)

func Load() *Config {
	mu.Lock()
	defer mu.Unlock()

	if cfg != nil {
		return cfg
	}

	cfg = &Config{
		Exclusions:          []string{".git", "node_modules", "vendor", "frontend"},
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
		_ = json.Unmarshal(data, cfg)
	}

	if cfg.Lang == "" {
		cfg.Lang = "ru"
	}
	return cfg
}

func Save(newCfg *Config) error {
	mu.Lock()
	defer mu.Unlock()
	cfg = newCfg
	return saveLocked(newCfg)
}

func saveLocked(c *Config) error {
	data, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(cfgPath, data, 0644)
}

func (c *Config) IsExcluded(path string, root string) bool {
	// Normalize both to handle Windows drive casing
	root = strings.ToLower(filepath.ToSlash(root))
	path = strings.ToLower(filepath.ToSlash(path))

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return false
	}

	cleanPath := filepath.ToSlash(rel)
	base := filepath.Base(cleanPath)

	for _, ex := range c.Exclusions {
		segments := strings.Split(cleanPath, "/")
		for _, s := range segments {
			if s == ex {
				return true
			}
		}
	}
	// Simple mask check (glob)
	for _, mask := range c.Masks {
		match, _ := filepath.Match(mask, base)
		if match {
			return true
		}
	}
	return false
}
