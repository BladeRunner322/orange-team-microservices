package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	grpcclient "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/client"
	grpcserver "github.com/BladeRunner322/orange-team-microservices/pkg/grpc/server"
	"github.com/BladeRunner322/orange-team-microservices/pkg/health"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/pkg/pgconfig"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/config"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/infrastructure/clients"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/infrastructure/postgres_repo"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/interfaces/workoutsgrpc"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer      *grpc.Server
	listener        net.Listener
	logger          *logger.Logger
	config          config.Config
	pool            *postgres.PgxPool
	httpServer      *http.Server
	profilesClient  *clients.ProfilesClient
	exercisesClient *clients.ExercisesClient
}

func New(cfg config.Config, log *logger.Logger) (*App, error) {
	ctx := context.Background()

	// 1. Postgres
	pgCfg := pgconfig.MustLoad()
	pool, err := postgres.NewPgxPool(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	log.Info("postgres connection pool created")

	// 2. Clients
	profilesClient, err := clients.NewProfilesClient(ctx, grpcclient.Config{
		Target:  cfg.ProfilesGRPCAddr,
		TLSMode: grpcclient.TLSModeInsecure,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("create profiles client: %w", err)
	}
	exercisesClient, err := clients.NewExercisesClient(ctx, grpcclient.Config{
		Target:  cfg.ExercisesGRPCAddr,
		TLSMode: grpcclient.TLSModeInsecure,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("create exercises client: %w", err)
	}

	// 3. Repositories
	workoutRepo := postgres_repo.NewWorkoutRepository(pool)
	exerciseRepo := postgres_repo.NewWorkoutExerciseRepository(pool)

	// 4. Usecases
	createWorkoutUC := usecases.NewCreateWorkout(workoutRepo, profilesClient, log)
	getWorkoutUC := usecases.NewGetWorkout(workoutRepo, log)
	getWorkoutsUC := usecases.NewGetWorkouts(workoutRepo, log)
	patchWorkoutUC := usecases.NewPatchWorkout(workoutRepo, log)
	deleteWorkoutUC := usecases.NewDeleteWorkout(workoutRepo, log)
	createWorkoutExerciseUC := usecases.NewCreateWorkoutExercise(workoutRepo, exerciseRepo, exercisesClient, log)
	getWorkoutExercisesUC := usecases.NewGetWorkoutExercises(workoutRepo, exerciseRepo, log)
	patchWorkoutExerciseUC := usecases.NewPatchWorkoutExercise(workoutRepo, exerciseRepo, exercisesClient, log)
	deleteWorkoutExerciseUC := usecases.NewDeleteWorkoutExercise(workoutRepo, exerciseRepo, log)
	getUserScoreUC := usecases.NewGetUserScore(workoutRepo, log)
	log.Info("use cases initialized")

	// 5. gRPC сервер
	s, err := grpcserver.New(grpcserver.Config{
		EnableTLS:        cfg.EnableTLS,
		TLSCertFile:      cfg.TLSCertFile,
		TLSKeyFile:       cfg.TLSKeyFile,
		EnableReflection: cfg.EnableReflection,
		WithUserID:       true,
		Logger:           log,
	})
	if err != nil {
		return nil, fmt.Errorf("create grpc server: %w", err)
	}

	workouts.RegisterWorkoutsServiceServer(s, workoutsgrpc.NewServer(
		createWorkoutUC, getWorkoutUC, getWorkoutsUC, patchWorkoutUC, deleteWorkoutUC,
		createWorkoutExerciseUC, getWorkoutExercisesUC, patchWorkoutExerciseUC, deleteWorkoutExerciseUC,
		getUserScoreUC,
	))
	log.Info("gRPC server registered")

	// 6. Listener
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	log.Info("gRPC listener created", "addr", cfg.GRPCPort)

	// 7. HTTP сервер
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.HealthHandler("workouts"))
	healthMux.HandleFunc("/ready", health.ReadyHandler(map[string]health.Check{
		"postgres":  health.PostgresCheck(pool),
		"profiles":  profilesClient.IsHealthy,
		"exercises": exercisesClient.IsHealthy,
	}))
	healthMux.Handle("/metrics", health.MetricsHandler())

	httpSrv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	go func() {
		log.Info("HTTP server listening", "addr", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed", "error", err)
		}
	}()

	return &App{
		grpcServer:      s,
		listener:        lis,
		logger:          log,
		config:          cfg,
		pool:            pool,
		httpServer:      httpSrv,
		profilesClient:  profilesClient,
		exercisesClient: exercisesClient,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("Workouts service listening", "addr", a.config.GRPCPort)
	return a.grpcServer.Serve(a.listener)
}

func (a *App) GracefulStop() { a.grpcServer.GracefulStop() }
func (a *App) Stop()         { a.grpcServer.Stop() }

func (a *App) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP server shutdown error", "error", err)
		}
	}
	if a.profilesClient != nil {
		a.profilesClient.Close()
	}
	if a.exercisesClient != nil {
		a.exercisesClient.Close()
	}
	if a.pool != nil {
		a.pool.Close()
	}
	a.logger.Info("all resources closed")
}
