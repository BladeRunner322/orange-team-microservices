# Сборка бинарника
FROM golang:1.26.7-bookworm AS builder

WORKDIR /app

# Копируем зависимости и исходники
COPY go.mod go.sum ./
COPY vendor /app/vendor
COPY . .

# Сборка с vendor (используем vendor для быстрой сборки)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /gateway ./services/gateway/cmd

# Финальный лёгкий образ
FROM alpine:3.23

# Устанавливаем полезные утилиты (curl для healthcheck, tzdata для времени)
RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

# Копируем бинарник
COPY --from=builder /gateway /app/gateway

ENV TZ=UTC

# Gateway слушает только HTTP (порт 8080)
EXPOSE 8080

CMD ["/app/gateway"]