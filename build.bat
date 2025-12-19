@echo off
echo === Starting Goruner Build Process ===

REM Check if wails is installed
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Wails CLI not found. Please install it first.
    exit /b 1
)

echo [1/3] Tidying Go modules...
go mod tidy


echo [3/3] Building Wails application...
wails build

if %errorlevel% equ 0 (
    echo === Build Successful! Executable is in build/bin ===
) else (
    echo === Build Failed! ===
    exit /b 1
)

pause