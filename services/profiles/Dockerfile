# Сборка бинарника
FROM golang:1.26.7-bookworm AS builder

WORKDIR /app

# Зависимости
COPY go.mod go.sum ./
COPY vendor /app/vendor
COPY . .

# Сборка с vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /profiles ./services/profiles/cmd

# Финальный образ
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /profiles /app/profiles

ENV TZ=UTC

EXPOSE 50052 8080

CMD ["/app/profiles"]
