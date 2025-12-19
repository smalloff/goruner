package config

import (
	"os"
	"testing"
)

func TestConfig_IsExcluded(t *testing.T) {
	c := &Config{
		Exclusions: []string{".git", "vendor"},
		Masks:      []string{"*.tmp"},
	}
	root := "/project"

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/.git/config", true},
		{"/project/src/main.go", false},
		{"/project/vendor/pkg/a.go", true},
		{"/project/test.tmp", true},
		{"/project/subdir/test.tmp", true},
	}

	for _, tt := range tests {
		if got := c.IsExcluded(tt.path, root); got != tt.expected {
			t.Errorf("IsExcluded(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	cfgPath = "test_config.json"
	defer os.Remove(cfgPath)

	orig := Load()
	orig.Lang = "en"
	orig.AlwaysOnTop = true

	err := Save(orig)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	instance = nil // Reset singleton
	loaded := Load()

	if loaded.Lang != "en" || !loaded.AlwaysOnTop {
		t.Errorf("Loaded config mismatch: %+v", loaded)
	}
}
