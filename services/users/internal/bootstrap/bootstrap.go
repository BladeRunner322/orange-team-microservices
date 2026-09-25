// Package bootstrap собирает зависимости Users-сервиса и запускает компоненты.
//
// Собирает:
//   - подключение к PostgreSQL
//   - репозиторий профилей
//   - usecases (GetMyProfile, PatchMyProfile, DeleteMyProfile, GetProfile)
//   - gRPC-сервер с интерсепторами (logging, metrics, recovery, UserIDServer)
//   - HTTP-сервер для /health и /metrics
//
// По образцу services/auth/internal/bootstrap/bootstrap.go
package bootstrap

// TODO: App struct + New() + Run() + GracefulStop() + Close()
