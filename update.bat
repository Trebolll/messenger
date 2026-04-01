@echo off
title Messenger - Update

cd /d "%~dp0"

echo Stopping services...
docker compose -f docker-compose.local.yml down

echo Rebuilding images...
docker compose -f docker-compose.local.yml build --no-cache

echo Starting updated services...
docker compose -f docker-compose.local.yml up -d

echo Update complete!
start "" "https://www.lambdahub.ru"
exit /b 0