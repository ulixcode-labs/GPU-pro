@echo off
REM GPU Pro - Quick Start Script (Windows)
REM This script makes it easy to build and run GPU Pro with a single command

setlocal enabledelayedexpansion

REM Start main execution
goto main

REM ================================================
REM Print messages
REM ================================================
:info
echo [INFO] %~1
goto :eof

:success
echo [SUCCESS] %~1
goto :eof

:error
echo [ERROR] %~1
goto :eof

:warning
echo [WARNING] %~1
goto :eof

REM ================================================
REM Check if Go is installed
REM ================================================
:check_go
where go >nul 2>nul
if %errorlevel% neq 0 (
    call :error "Go is not installed!"
    echo.
    echo Please install Go 1.24+ from: https://golang.org/dl/
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do (
    set GO_VERSION=%%i
    set GO_VERSION=!GO_VERSION:go=!
)
call :success "Go !GO_VERSION! detected"
goto :eof

REM ================================================
REM Build Web UI
REM ================================================
:build_webui
call :info "Building Web UI version..."
go build -ldflags="-s -w" -o gpu-pro.exe .
if %errorlevel% neq 0 (
    call :error "Web UI build failed"
    exit /b 1
)
call :success "Web UI build successful: gpu-pro.exe"
goto :eof

REM ================================================
REM Build Terminal UI
REM ================================================
:build_tui
call :info "Building Terminal UI version..."
go build -ldflags="-s -w" -o gpu-pro-cli.exe ./cmd/gpu-pro-cli
if %errorlevel% neq 0 (
    call :error "Terminal UI build failed"
    exit /b 1
)
call :success "Terminal UI build successful: gpu-pro-cli.exe"
goto :eof

REM ================================================
REM Check if binaries exist
REM ================================================
:check_binaries
set WEBUI_EXISTS=false
set TUI_EXISTS=false

if exist "gpu-pro.exe" set WEBUI_EXISTS=true
if exist "gpu-pro-cli.exe" set TUI_EXISTS=true
goto :eof

REM ================================================
REM Run Web UI
REM ================================================
:run_webui
echo.
call :info "Starting GPU Pro Web UI..."
echo.
echo ==================================================
echo   Web Dashboard: http://localhost:8889
echo ==================================================
echo.
call :info "Press Ctrl+C to stop the server"
echo.

gpu-pro.exe
goto :eof

REM ================================================
REM Run Terminal UI
REM ================================================
:run_tui
echo.
call :info "Starting GPU Pro Terminal UI..."
echo.

gpu-pro-cli.exe
goto :eof

REM ================================================
REM Clean binaries
REM ================================================
:clean_binaries
call :info "Cleaning built binaries..."
if exist "gpu-pro.exe" del /f gpu-pro.exe
if exist "gpu-pro-cli.exe" del /f gpu-pro-cli.exe
call :success "Clean complete"
goto :eof

REM ================================================
REM Main function
REM ================================================
:main
cls
echo.
echo ==================================================
echo            GPU Pro - Quick Start
echo         Master Your AI Workflow
echo ==================================================
echo.

REM Check Go installation
call :check_go
if %errorlevel% neq 0 (
    pause
    exit /b 1
)
echo.

REM Parse command line arguments for direct execution
if "%~1"=="web" goto direct_web
if "%~1"=="webui" goto direct_web
if "%~1"=="1" goto direct_web
if "%~1"=="tui" goto direct_tui
if "%~1"=="cli" goto direct_tui
if "%~1"=="2" goto direct_tui
if "%~1"=="build" goto direct_build
if "%~1"=="3" goto direct_build
if "%~1"=="clean" goto direct_clean
if "%~1"=="4" goto direct_clean
if "%~1"=="quit" goto direct_quit
if "%~1"=="q" goto direct_quit
if "%~1"=="Q" goto direct_quit

goto interactive_menu

:direct_web
call :check_binaries
if "!WEBUI_EXISTS!"=="false" (
    call :build_webui
    if %errorlevel% neq 0 (
        pause
        exit /b 1
    )
) else (
    call :success "Using existing binary: gpu-pro.exe"
)
call :run_webui
exit /b 0

:direct_tui
call :check_binaries
if "!TUI_EXISTS!"=="false" (
    call :build_tui
    if %errorlevel% neq 0 (
        pause
        exit /b 1
    )
) else (
    call :success "Using existing binary: gpu-pro-cli.exe"
)
call :run_tui
exit /b 0

:direct_build
call :build_webui
call :build_tui
echo.
call :success "Build complete! Use 'start.bat' to run."
echo.
pause
exit /b 0

:direct_clean
call :clean_binaries
pause
exit /b 0

:direct_quit
echo.
call :info "Goodbye!"
echo.
exit /b 0

REM ================================================
REM Interactive menu
REM ================================================
:interactive_menu
call :check_binaries

if "!WEBUI_EXISTS!"=="true" (
    call :success "Found existing Web UI binary"
)

if "!TUI_EXISTS!"=="true" (
    call :success "Found existing Terminal UI binary"
)

:menu_loop
echo.
echo Choose an option:
echo.
echo   1) Web UI      - Beautiful web dashboard (http://localhost:8889)
echo   2) Terminal UI - Elegant TUI for terminal sessions
echo   3) Build both  - Just build, don't run
echo   4) Clean       - Remove built binaries
echo   q) Quit
echo.
set /p choice=Enter your choice [1-4/q]:

if "%choice%"=="1" goto choice_1
if "%choice%"=="2" goto choice_2
if "%choice%"=="3" goto choice_3
if "%choice%"=="4" goto choice_4
if /i "%choice%"=="q" goto choice_quit

echo.
call :error "Invalid choice. Please enter 1, 2, 3, 4, or q"
goto menu_loop

:choice_1
if "!WEBUI_EXISTS!"=="false" (
    echo.
    call :build_webui
    if %errorlevel% neq 0 goto menu_loop
    set WEBUI_EXISTS=true
)
call :run_webui
goto end

:choice_2
if "!TUI_EXISTS!"=="false" (
    echo.
    call :build_tui
    if %errorlevel% neq 0 goto menu_loop
    set TUI_EXISTS=true
)
call :run_tui
goto end

:choice_3
echo.
call :build_webui
echo.
call :build_tui
echo.
call :success "Build complete! Run 'start.bat' again to start."
echo.
pause
exit /b 0

:choice_4
echo.
call :clean_binaries
echo.
pause
exit /b 0

:choice_quit
echo.
call :info "Goodbye!"
echo.
exit /b 0

:end
exit /b 0
