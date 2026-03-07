# Build stage
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые инструменты для сборки (если потребуются)
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Копируем исполняемый файл
COPY --from=builder /app/main .

# Копируем статику и шаблоны
COPY --from=builder /app/web ./web

# Копируем миграции (если используются в коде)
COPY --from=builder /app/internal/db/migration ./internal/db/migration

# Пробрасываем порт
EXPOSE 8080

# Запуск
CMD ["./main"]
