// Package config загружает конфигурацию Users-сервиса из переменных окружения.
//
// Переменные:
//   - GRPC_PORT          — порт gRPC-сервера
//   - HTTP_PORT          — порт HTTP-сервера (health, metrics)
//   - POSTGRES_*         — параметры подключения к PostgreSQL
//   - ENABLE_TLS         — использовать ли TLS для gRPC
//   - TLS_CERT_FILE      — путь к сертификату
//   - TLS_KEY_FILE       — путь к приватному ключу
//   - ENABLE_REFLECTION  — включить gRPC reflection для grpcurl
package config

// TODO: Config struct + Load() + MustLoad() — по образцу services/auth/config/config.go
