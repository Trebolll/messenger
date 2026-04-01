@echo off
title Messenger - Stop

cd /d "%~dp0"

docker compose -f docker-compose.local.yml down

echo Messenger stopped.
timeout /t 2 /nobreak >nul
exit /b 0