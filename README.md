# Messenger

A simple Go application demonstrating basic syntax and debugging features in GoLand.

## Features
- Prints a welcome message.
- Iterates through a loop and performs basic arithmetic.
- Includes tips for GoLand IDE integration.

## Requirements
- **Go**: 1.25 or higher

## Getting Started

### Running the Application
To run the application, use the following command in your terminal:
```bash
go run main.go
```

### Debugging
You can set breakpoints in `main.go` and start a debugging session via your IDE (e.g., GoLand).

## Project Structure
- `main.go`: Main logic of the application.
- `go.mod`: Module dependencies and Go version.


cat ~/.ssh/id_ed25519Значит в репо всё ещё старый docker-compose.yml с DB_PASSWORD: postgres. Проверь локально в репо — есть ли там DB_PASSWORD в секции бэкенда?
docker exec -it messenger-db psql -U postgres -d messenger


.\ngrok.exe start --all --config=ngrok.yml
docker-compose up --build -d backend
docker-compose logs -f backend
docker compose -f docker-compose.local.yml up -d --build

ssh root@193.233.103.158



авторизация гит на сервере

ssh -p 9871 root@193.233.103.158
cd /opt/messenger
git remote set-url origin git@github.com:Trebolll/messenger.git
git fetch origin
git reset --hard origin/master

обновить код на сервере
cd /opt/messenger
git pull origin master
docker compose build backend
docker compose up -d --force-recreate backend