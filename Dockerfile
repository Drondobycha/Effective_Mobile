# Базовый образ для сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o migrator ./cmd/migrator/main.go
RUN go build -o server ./cmd/server/main.go

# Финальный образ для сервера
FROM alpine:latest AS server

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./app/migrations

EXPOSE 8080

CMD ["./server"]

# Финальный образ для мигратора
FROM alpine:latest AS migrator

WORKDIR /app

COPY --from=builder /app/migrator .
COPY --from=builder /app/migrations ./app/migrations

CMD ["./migrator"]