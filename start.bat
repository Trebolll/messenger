@echo off
title Messenger - Start

cd /d "%~dp0"

docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo Docker is not running. Starting Docker Desktop...
    start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    echo Waiting for Docker...
    set /a count=0
    :wait_docker
    timeout /t 5 /nobreak >nul
    docker info >nul 2>&1
    if %errorlevel% == 0 goto :docker_ok
    set /a count+=1
    if %count% lss 24 goto :wait_docker
    echo ERROR: Docker failed to start. Open Docker Desktop manually and try again.
    pause
    exit /b 1
)

:docker_ok
echo Docker is ready.
echo Starting services...
docker compose -f docker-compose.local.yml up -d --build

if %errorlevel% neq 0 (
    echo ERROR: Failed to start. Check logs: docker compose logs
    pause
    exit /b 1
)

echo.
echo Messenger started!
echo Opening https://www.lambdahub.ru ...

timeout /t 2 /nobreak >nul
start "" "https://www.lambdahub.ru"

exit /b 0