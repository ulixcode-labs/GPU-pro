@echo off
REM GPU Pro - Windows Runner Script
REM This script builds and runs GPU Pro on Windows

echo Building GPU Pro (Windows - with nvidia-smi GPU monitoring)...
echo Note: GPU monitoring uses nvidia-smi (requires NVIDIA drivers)
echo.
go build -ldflags "-s -w" -o gpu-pro.exe .

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b %errorlevel%
)

echo.
echo Starting GPU Pro (Web UI)...
echo Web Dashboard: http://localhost:8889
echo.
gpu-pro.exe
