package main

import (
	"context"
	"fmt"
	"goruner/internal/config"
	"goruner/internal/notifier"
	"goruner/internal/tester"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App defines the main Wails application controller.
type App struct {
	ctx      context.Context
	watcher  *fsnotify.Watcher
	mu       sync.Mutex
	isPaused bool
}

// NewApp initializes a new App instance.
func NewApp() *App {
	return &App{isPaused: false}
}

// startup is called by Wails when the application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initWatcher()

	cfg := config.Load()
	if cfg.AlwaysOnTop {
		wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
	}
}

// initWatcher configures the filesystem observer.
func (a *App) initWatcher() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	a.watcher = w

	go a.watchLoop()
	a.refreshWatcher()
}

// watchLoop processes filesystem events.
func (a *App) watchLoop() {
	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}
		
			a.mu.Lock()
			paused := a.isPaused
			a.mu.Unlock()

			if paused {
				continue
			}

			cwd, _ := os.Getwd()
			cfg := config.Load()

			// Auto-watch new directories
			if event.Op&fsnotify.Create != 0 {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() && !cfg.IsExcluded(event.Name, cwd) {
					_ = a.watcher.Add(event.Name)
				}
			}

			// Trigger tests on Go file modifications
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
				if strings.HasSuffix(strings.ToLower(event.Name), ".go") && 
				   cfg.AutoWatch && !cfg.IsExcluded(event.Name, cwd) {
					wailsRuntime.EventsEmit(a.ctx, "trigger_test", "File changed: "+filepath.Base(event.Name))
				}
			}
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("[Watcher] Error: %v\n", err)
		}
	}
}

// RunTests executes Go tests and processes results for the UI.
func (a *App) RunTests() string {
	cfg := config.Load()
	cwd, _ := os.Getwd()

	pkgs, err := tester.DiscoverTests(cwd)
	if err != nil {
		return fmt.Sprintf("Discovery Error: %v", err)
	}
	
	out, err := tester.RunTests(a.ctx, cwd, pkgs, cfg.ShowPassed, cfg.Lang)
	if err != nil {
		return fmt.Sprintf("Execution Error: %v\nOutput: %s", err, out)
	}

	isError := strings.Contains(out, "FAIL") || strings.Contains(out, "Error")

	if cfg.ShowNotifications {
		if !cfg.NotifyOnlyOnFailure || isError {
			a.sendNotify(out, isError, cfg.Lang)
		}
	}

	if isError && cfg.AutoCopyErrors {
		a.copyErrorsToClipboard(out)
	}

	return out
}

func (a *App) sendNotify(output string, isError bool, lang string) {
	title := "Go Test Runner"
	var msg string
	if lang == "ru" {
		if isError { msg = "❌ Тесты провалены!" } else { msg = "✅ Все тесты пройдены!" }
	} else {
		if isError { msg = "❌ Tests failed!" } else { msg = "✅ All tests passed!" }
	}
	notifier.Notify(title, msg)
}

func (a *App) copyErrorsToClipboard(output string) {
	lines := strings.Split(output, "\n")
	var errLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}

		isError := strings.HasPrefix(trimmed, "#") ||
			strings.Contains(l, ".go:") ||
			strings.Contains(l, "FAIL") ||
			strings.Contains(strings.ToLower(l), "error:") ||
			strings.Contains(l, "panic:")

		if isError {
			errLines = append(errLines, l)
		}
	}
	if len(errLines) > 0 {
		_ = wailsRuntime.ClipboardSetText(a.ctx, strings.Join(errLines, "\n"))
	}
}

func (a *App) refreshWatcher() {
	cfg := config.Load()
	cwd, _ := os.Getwd()
	_ = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() { return nil }
		if cfg.IsExcluded(path, cwd) { return filepath.SkipDir }
		_ = a.watcher.Add(path)
		return nil
	})
}

// --- Wails Bindings ---

// GetDiscoveredTests returns a list of packages that contain tests.
func (a *App) GetDiscoveredTests() []string {
	cwd, _ := os.Getwd()
	pkgs, _ := tester.DiscoverTests(cwd)
	return pkgs
}

// TogglePause switches the file watcher's active state.
func (a *App) TogglePause() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isPaused = !a.isPaused
	return a.isPaused
}

// GetConfig retrieves the current configuration instance.
func (a *App) GetConfig() *config.Config { return config.Load() }

// SaveConfig updates and persists application settings.
func (a *App) SaveConfig(cfg config.Config) {
	_ = config.Save(&cfg)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, cfg.AlwaysOnTop)
}

// Quit terminates the application gracefully.
func (a *App) Quit() {
	wailsRuntime.Quit(a.ctx)
}
