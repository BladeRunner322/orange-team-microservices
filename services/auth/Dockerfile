FROM golang:1.26.7-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /auth ./services/auth/cmd

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /auth /app/auth

ENV TZ=UTC

EXPOSE 50051
EXPOSE 8080

CMD ["/app/auth"]