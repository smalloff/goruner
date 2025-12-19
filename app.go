package main

import (
	"context"
	"fmt"
	"goruner/internal/config"
	"goruner/internal/tester"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	watcher  *fsnotify.Watcher
	mu       sync.Mutex
	isPaused bool
}

func NewApp() *App {
	return &App{isPaused: false}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initWatcher()
	
	cfg := config.Load()
	if cfg.AlwaysOnTop {
		wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
	}

	// Запуск системного трея в отдельном потоке
	go systray.Run(a.onTrayReady, a.onTrayExit)
}

func (a *App) initWatcher() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	a.watcher = w

	go func() {
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
				
				if event.Op&fsnotify.Create != 0 {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						cfg := config.Load()
						if !cfg.IsExcluded(event.Name, cwd) {
							_ = a.watcher.Add(event.Name)
						}
					}
				}
				
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					cfg := config.Load()
					if strings.HasSuffix(strings.ToLower(event.Name), ".go") && cfg.AutoWatch && !cfg.IsExcluded(event.Name, cwd) {
						fmt.Printf("[Watcher] Triggering test: %s (Op: %v)\n", event.Name, event.Op)
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
	}()

	a.refreshWatcher()
}

func (a *App) refreshWatcher() {
	cfg := config.Load()
	cwd, _ := os.Getwd()
	
	_ = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if cfg.IsExcluded(path, cwd) {
			return filepath.SkipDir
		}
		_ = a.watcher.Add(path)
		return nil
	})
}

func (a *App) RunTests() string {
	cfg := config.Load()
	cwd, _ := os.Getwd()
	out, err := tester.RunTests(a.ctx, cwd, cfg.ShowPassed, cfg.Lang)
	if err != nil {
		return fmt.Sprintf("Execution Error: %v\nOutput: %s", err, out)
	}

	isError := strings.Contains(out, "FAIL") || strings.Contains(out, "Error")
	
	if cfg.ShowNotifications {
		if !cfg.NotifyOnlyOnFailure || isError {
			a.sendNotification(out, cfg.Lang)
		}
	}
	
	if isError && cfg.AutoCopyErrors {
		_ = wailsRuntime.ClipboardSetText(a.ctx, out)
	}
	
	return out
}

func (a *App) sendNotification(output string, lang string) {
	if runtime.GOOS != "windows" {
		return
	}

	title := "Go Test Runner"
	msg := ""
	isError := strings.Contains(output, "FAIL") || strings.Contains(output, "Error")

	if lang == "ru" {
		if isError {
			msg = "❌ Тесты провалены! Обнаружены ошибки."
		} else {
			msg = "✅ Все тесты пройдены успешно!"
		}
	} else {
		if isError {
			msg = "❌ Tests failed! Errors detected."
		} else {
			msg = "✅ All tests passed successfully!"
		}
	}

		notification := fmt.Sprintf("[void][System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms'); $obj = New-Object System.Windows.Forms.NotifyIcon; $obj.Icon = [System.Drawing.SystemIcons]::Information; $obj.BalloonTipTitle = '%s'; $obj.BalloonTipText = '%s'; $obj.Visible = $true; $obj.ShowBalloonTip(5000);", title, msg)
		cmd := exec.Command("powershell", "-NoProfile", "-Command", notification)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = cmd.Run()
	}

func (a *App) GetDiscoveredTests() []string {
	cwd, _ := os.Getwd()
	pkgs, _ := tester.DiscoverTests(cwd)
	return pkgs
}

func (a *App) TogglePause() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isPaused = !a.isPaused
	return a.isPaused
}

func (a *App) GetConfig() *config.Config {
	return config.Load()
}

func (a *App) SaveConfig(cfg config.Config) {
	_ = config.Save(&cfg)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, cfg.AlwaysOnTop)
}

// onDomReady вызывается, когда фронтенд готов. Мы подписываемся на события окна.
func (a *App) onDomReady(ctx context.Context) {
	wailsRuntime.EventsOn(ctx, "wails:window:minimize", func(optionalData ...interface{}) {
		cfg := config.Load()
		if cfg.MinimizeToTray {
			wailsRuntime.WindowHide(ctx)
		}
	})
}

// RestoreWindow восстанавливает окно из трея
func (a *App) RestoreWindow() {
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowUnminimise(a.ctx)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, config.Load().AlwaysOnTop)
}

// onTrayReady настраивает меню трея
func (a *App) onTrayReady() {
	// Установка иконки (пустая иконка или загруженная)
	systray.SetTitle("Go Runner")
	systray.SetTooltip("Go Test Runner")

	mOpen := systray.AddMenuItem("Открыть", "Восстановить окно")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Выход", "Закрыть приложение")

	for {
		select {
		case <-mOpen.ClickedCh:
			a.RestoreWindow()
		case <-mQuit.ClickedCh:
			a.Quit()
		}
	}
}

func (a *App) onTrayExit() {
	// Очистка при выходе
}

// Quit завершает работу приложения
func (a *App) Quit() {
	systray.Quit()
	wailsRuntime.Quit(a.ctx)
}