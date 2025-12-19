@echo off
setlocal
echo === Starting Goruner Build Process ===

REM Check for Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go not found in PATH.
    exit /b 1
)

REM Check for Wails CLI
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Wails CLI not found. Please install: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    exit /b 1
)

echo [1/3] Tidying Go modules...
go mod tidy

echo [2/3] Running backend tests...
go test ./internal/...
if %errorlevel% neq 0 (
    echo ERROR: Tests failed. Build aborted.
    exit /b 1
)

echo [3/3] Building Wails application (Production mode)...
wails build -ldflags="-s -w"

if %errorlevel% equ 0 (
    echo === Build Successful! ===
) else (
    echo === Build Failed! ===
    exit /b 1
)

pause
endlocal