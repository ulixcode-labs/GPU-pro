@echo off
REM GPU Pro CLI - Windows Runner Script
REM This script builds and runs GPU Pro CLI/TUI on Windows

echo Building GPU Pro CLI (Windows - with nvidia-smi GPU monitoring)...
echo Note: GPU monitoring uses nvidia-smi (requires NVIDIA drivers)
echo.
go build -ldflags "-s -w" -o gpu-pro-cli.exe ./cmd/gpu-pro-cli

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b %errorlevel%
)

echo.
echo Starting GPU Pro (CLI/TUI)...
echo.
gpu-pro-cli.exe
