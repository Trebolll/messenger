# ===== ЭТАП 1: Сборка =====
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o messenger ./cmd/server

# ===== ЭТАП 2: Запуск =====
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/messenger .
COPY web ./web

EXPOSE 8080

ENTRYPOINT ["./messenger"]
