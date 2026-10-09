package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"

	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	"github.com/BladeRunner322/orange-team-microservices/pkg/health"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/ratelimit"
	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/config"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/infrastructure/clients"
	authhandlers "github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/handlers/auth"
	exerciseshandlers "github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/handlers/exercises"
	profileshandlers "github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/handlers/profiles"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/handlers/proxy"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/middleware"
)

// App объединяет все компоненты Gateway.
type App struct {
	httpServer      *http.Server
	logger          *logger.Logger
	authClient      *clients.AuthClient
	profilesClient  *clients.ProfilesClient
	exercisesClient *clients.ExercisesClient
	redisClient     *redis.Client
}

// New создаёт экземпляр App, собирает все зависимости.
func New(cfg config.Config, log *logger.Logger) (*App, error) {
	ctx := context.Background()

	// 1. gRPC-клиент к Auth Service
	authClient, err := clients.NewAuthClient(ctx, buildGRPCConfig(cfg.AuthGRPCAddr, cfg))
	if err != nil {
		return nil, fmt.Errorf("create auth client: %w", err)
	}
	log.Info("auth gRPC client created",
		"addr", cfg.AuthGRPCAddr,
		"timeout", cfg.Timeout,
		"cb_enabled", cfg.CBEnabled,
		"retry_enabled", cfg.RetryEnabled,
	)

	// 2. gRPC-клиент к Profiles Service
	profilesClient, err := clients.NewProfilesClient(ctx, buildGRPCConfig(cfg.ProfilesGRPCAddr, cfg))
	if err != nil {
		return nil, fmt.Errorf("create profiles client: %w", err)
	}
	log.Info("profiles gRPC client created",
		"addr", cfg.ProfilesGRPCAddr,
		"timeout", cfg.Timeout,
		"cb_enabled", cfg.CBEnabled,
		"retry_enabled", cfg.RetryEnabled,
	)

	// 3. gRPC-клиент к Exercises Service
	exercisesClient, err := clients.NewExercisesClient(ctx, buildGRPCConfig(cfg.ExercisesGRPCAddr, cfg))
	if err != nil {
		return nil, fmt.Errorf("create exercises client: %w", err)
	}
	log.Info("exercises gRPC client created",
		"addr", cfg.ExercisesGRPCAddr,
		"timeout", cfg.Timeout,
		"cb_enabled", cfg.CBEnabled,
		"retry_enabled", cfg.RetryEnabled,
	)

	// 4. Redis-клиент для rate limiting
	redisClient, err := redis.NewClient(ctx, redis.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}
	log.Info("redis client created", "addr", cfg.RedisAddr)

	// 5. Rate limiter
	limiter := ratelimit.NewLimiter(redisClient.Client)

	// 6. Trusted proxies для определения IP клиента
	trustedProxies, err := middleware.ParseTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return nil, fmt.Errorf("parse trusted proxies: %w", err)
	}
	if len(trustedProxies) > 0 {
		log.Info("trusted proxies configured", "count", len(trustedProxies))
	} else {
		log.Info("trusted proxies not configured — X-Forwarded-For will be ignored")
	}

	// 7. Readiness checks (Redis + gRPC-клиенты)
	readinessChecks := map[string]health.Check{
		"redis":     health.RedisCheck(redisClient),
		"auth":      authClient.IsHealthy,
		"profiles":  profilesClient.IsHealthy,
		"exercises": exercisesClient.IsHealthy,
	}

	// 8. Роутер
	r := chi.NewRouter()
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(chimid.Recoverer)
	r.Use(middleware.HTTPMetricsMiddleware)

	// 9. Служебные эндпоинты (без rate limit).
	//
	// /health, /ready и /metrics опрашиваются инфраструктурой (Docker
	// healthcheck, Prometheus, внешний LB) с фиксированной частотой.
	// Rate limit на них смысла не имеет: клиента за ними нет, а
	// лишние обращения к Redis замедляют healthcheck.
	r.Group(func(r chi.Router) {
		r.Get("/health", health.HealthHandler("gateway"))
		r.Get("/ready", health.ReadyHandler(readinessChecks))
		r.Handle("/metrics", health.MetricsHandler())
	})

	// 10. Публичные маршруты (rate limit по IP/email).
	rateLimitPublic := middleware.RateLimitMiddleware(limiter, trustedProxies, cfg.RateLimit, log)

	r.Group(func(r chi.Router) {
		r.Use(rateLimitPublic)

		r.Post("/register", authhandlers.RegisterHandler(authClient))
		r.Post("/login", authhandlers.LoginHandler(authClient))
		r.Post("/refresh", authhandlers.RefreshHandler(authClient))
		r.Post("/logout", authhandlers.LogoutHandler(authClient))
	})

	// 11. Защищённые маршруты (обычные пользователи, без RBAC)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authClient))
		r.Use(middleware.RateLimitMiddleware(limiter, trustedProxies, cfg.RateLimit, log))

		// Users
		r.Get("/users/me", profileshandlers.GetUserHandler(profilesClient))
		r.Patch("/users/me", profileshandlers.PatchUserHandler(profilesClient))
		r.Delete("/users/me", profileshandlers.DeleteUserHandler(profilesClient))

		// Exercises (чтение — всем авторизованным)
		r.Get("/exercises", exerciseshandlers.GetExercisesHandler(exercisesClient))
		r.Get("/exercises/{exerciseId}", exerciseshandlers.GetExerciseHandler(exercisesClient))

		// Habits
		r.Get("/habits", proxy.Handler)
		r.Post("/habits", proxy.Handler)
		r.Post("/habits/{habitId}/complete", proxy.Handler)
		r.Delete("/habits/{habitId}", proxy.Handler)

		// Workouts
		r.Get("/workouts", proxy.Handler)
		r.Post("/workouts", proxy.Handler)
		r.Get("/workouts/{workoutId}", proxy.Handler)
		r.Patch("/workouts/{workoutId}", proxy.Handler)
		r.Delete("/workouts/{workoutId}", proxy.Handler)
		r.Post("/workouts/{workoutId}/exercises", proxy.Handler)
		r.Get("/workouts/{workoutId}/exercises", proxy.Handler)
		r.Patch("/workouts/{workoutId}/exercises/{exerciseId}", proxy.Handler)
		r.Delete("/workouts/{workoutId}/exercises/{exerciseId}", proxy.Handler)

		// Leaderboard
		r.Get("/leaderboard/daily", proxy.Handler)
		r.Get("/leaderboard/weekly", proxy.Handler)
		r.Get("/leaderboard/monthly", proxy.Handler)
	})

	// 12. Admin-only маршруты (RBAC: role == admin)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authClient))
		r.Use(middleware.RateLimitMiddleware(limiter, trustedProxies, cfg.RateLimit, log))
		r.Use(middleware.RequireRole("admin"))

		// Exercises (мутации — только admin)
		r.Post("/exercises", exerciseshandlers.CreateExerciseHandler(exercisesClient))
		r.Patch("/exercises/{exerciseId}", exerciseshandlers.PatchExerciseHandler(exercisesClient))
		r.Delete("/exercises/{exerciseId}", exerciseshandlers.DeleteExerciseHandler(exercisesClient))
	})

	// 13. HTTP-сервер
	httpSrv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Info("HTTP server configured", "port", cfg.HTTPPort)

	return &App{
		httpServer:      httpSrv,
		logger:          log,
		authClient:      authClient,
		profilesClient:  profilesClient,
		exercisesClient: exercisesClient,
		redisClient:     redisClient,
	}, nil
}

// Run запускает HTTP-сервер (блокирует выполнение).
func (a *App) Run() error {
	a.logger.Info("Gateway listening", "addr", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

// GracefulStop останавливает сервер с таймаутом 5 секунд.
func (a *App) GracefulStop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.httpServer.Shutdown(ctx)
}

// Close закрывает ресурсы.
func (a *App) Close() {
	if a.authClient != nil {
		a.authClient.Close()
	}

	if a.profilesClient != nil {
		a.profilesClient.Close()
	}

	if a.exercisesClient != nil {
		a.exercisesClient.Close()
	}

	if a.redisClient != nil {
		_ = a.redisClient.Close()
	}
}

// buildGRPCConfig собирает grpcclient.Config для одного downstream.
//
// CB/Retry-параметры общие для всех клиентов Gateway (одни
// env-переменные), Target — разный для каждого сервиса.
//
// Bootstrap передаёт результат в clients.NewXxxClient, который
// прокидывает его в pkg/grpc/client.New без изменений.
func buildGRPCConfig(target string, cfg config.Config) grpcclient.Config {
	return grpcclient.Config{
		Target:  target,
		TLSMode: grpcclient.TLSMode(cfg.GRPCClientTLSMode),
		Timeout: cfg.Timeout,

		// Circuit Breaker (ADR-020).
		CBEnabled:     cfg.CBEnabled,
		CBMaxRequests: cfg.CBMaxRequests,
		CBInterval:    cfg.CBInterval,
		CBTimeout:     cfg.CBTimeout,
		CBMinRequests: cfg.CBMinRequests,
		CBErrorRate:   cfg.CBErrorRate,

		// Retry (ADR-020).
		RetryEnabled:     cfg.RetryEnabled,
		RetryMaxAttempts: cfg.RetryMaxAttempts,
		RetryBaseDelay:   cfg.RetryBaseDelay,
		RetryMaxDelay:    cfg.RetryMaxDelay,
		RetryMethods:     cfg.RetryIdempotentMethods,
	}
}
