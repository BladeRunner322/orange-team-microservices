package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/config"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/infrastructure/clients"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/handlers"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/middleware"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
)

// App объединяет все компоненты Gateway
type App struct {
	httpServer *http.Server
	logger     *logger.Logger
	authClient *clients.AuthClient
}

// New создаёт экземпляр App, собирает все зависимости
func New(cfg config.Config, log *logger.Logger) (*App, error) {
	// 1. Создаём gRPC-клиент к Auth Service
	authClient, err := clients.NewAuthClient(cfg.AuthGRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("create auth client: %w", err)
	}
	log.Info("auth gRPC client created", "addr", cfg.AuthGRPCAddr)

	// 2. Создаём роутер
	r := chi.NewRouter()
	r.Use(middleware.RequestIDMiddleware)   // сначала request_id
	r.Use(middleware.LoggerMiddleware(log)) // логирование с полями (заменяет chi.Logger)
	r.Use(chimid.Recoverer)                 // восстановление после паники

	// 3. Публичные маршруты (не требуют токена)
	r.Get("/health", handlers.HealthHandler)
	r.Post("/register", handlers.RegisterHandler(authClient))
	r.Post("/login", handlers.LoginHandler(authClient))
	r.Post("/refresh", handlers.RefreshHandler(authClient))
	r.Post("/logout", handlers.LogoutHandler(authClient))

	// 4. Защищённые маршруты (требуют валидный JWT)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authClient))

		// Users
		r.Get("/users/me", handlers.GetUserHandler)
		r.Patch("/users/me", handlers.ProxyHandler)
		r.Delete("/users/me", handlers.ProxyHandler)

		// Exercises
		r.Get("/exercises", handlers.ProxyHandler)
		r.Post("/exercises", handlers.ProxyHandler) // admin only, позже добавим проверку ролей

		// Habits
		r.Get("/habits", handlers.ProxyHandler)
		r.Post("/habits", handlers.ProxyHandler)
		r.Post("/habits/{habitId}/complete", handlers.ProxyHandler)
		r.Delete("/habits/{habitId}", handlers.ProxyHandler)

		// Workouts
		r.Get("/workouts", handlers.ProxyHandler)
		r.Post("/workouts", handlers.ProxyHandler)
		r.Get("/workouts/{workoutId}", handlers.ProxyHandler)
		r.Patch("/workouts/{workoutId}", handlers.ProxyHandler)
		r.Delete("/workouts/{workoutId}", handlers.ProxyHandler)
		r.Post("/workouts/{workoutId}/exercises", handlers.ProxyHandler)
		r.Get("/workouts/{workoutId}/exercises", handlers.ProxyHandler)
		r.Patch("/workouts/{workoutId}/exercises/{exerciseId}", handlers.ProxyHandler)
		r.Delete("/workouts/{workoutId}/exercises/{exerciseId}", handlers.ProxyHandler)

		// Leaderboard
		r.Get("/leaderboard/daily", handlers.ProxyHandler)
		r.Get("/leaderboard/weekly", handlers.ProxyHandler)
		r.Get("/leaderboard/monthly", handlers.ProxyHandler)
	})

	// 5. HTTP-сервер
	httpSrv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Info("HTTP server configured", "port", cfg.HTTPPort)

	return &App{
		httpServer: httpSrv,
		logger:     log,
		authClient: authClient,
	}, nil
}

// Run запускает HTTP-сервер (блокирует выполнение)
func (a *App) Run() error {
	a.logger.Info("Gateway listening", "addr", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

// GracefulStop останавливает сервер с таймаутом 5 секунд
func (a *App) GracefulStop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.httpServer.Shutdown(ctx)
}

// Close закрывает ресурсы (gRPC-соединения)
func (a *App) Close() {
	if a.authClient != nil {
		a.authClient.Close()
	}
}
