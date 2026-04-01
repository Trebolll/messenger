@echo off
title Messenger - Prerequisites

echo Checking Docker...

docker --version >nul 2>&1
if %errorlevel% == 0 (
    echo Docker already installed.
    goto :check_running
)

echo Docker not found. Downloading Docker Desktop...

powershell -Command "& {[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://desktop.docker.com/win/main/amd64/Docker%%20Desktop%%20Installer.exe' -OutFile '%TEMP%\DockerDesktopInstaller.exe' -UseBasicParsing}"

if not exist "%TEMP%\DockerDesktopInstaller.exe" (
    echo ERROR: Failed to download Docker Desktop.
    echo Please install it manually from https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo Installing Docker Desktop...
"%TEMP%\DockerDesktopInstaller.exe" install --quiet --accept-license
del "%TEMP%\DockerDesktopInstaller.exe"

echo Docker Desktop installed.
echo IMPORTANT: Please restart your computer, then launch Messenger again.
pause
exit /b 0

:check_running
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo Starting Docker Desktop...
    start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    echo Waiting for Docker (up to 60 seconds)...
    set /a count=0
    :wait_loop
    timeout /t 5 /nobreak >nul
    docker info >nul 2>&1
    if %errorlevel% == 0 goto :docker_ready
    set /a count+=1
    if %count% lss 12 goto :wait_loop
    echo WARNING: Docker did not start in time. Please try launching Messenger again.
    exit /b 1
)

:docker_ready
echo Docker is ready.
exit /b 0