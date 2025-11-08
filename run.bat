@echo off
REM GPU Pro - Windows Runner Script
REM This script builds and runs GPU Pro on Windows

echo Building GPU Pro...
go build -ldflags "-s -w" -o gpu-pro.exe .

if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)

echo.
echo Starting GPU Pro (Web UI)...
echo.
gpu-pro.exe
