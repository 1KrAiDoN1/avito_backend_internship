package app

import (
	"context"
	"fmt"
	"internship/internal/config"
	httpserver "internship/internal/http-server"
	"internship/internal/http-server/handler"
	"internship/internal/repository/postgres"
	"internship/internal/service"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func Run(ctx context.Context, log *zap.Logger, config config.ServiceConfig) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	storage, err := postgres.NewDatabase(ctx, config.DbConfig.DBConn)
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Connected to database", zap.String("dsn", config.DbConfig.DBConn))

	dbpool := storage.GetPool()
	defer func() {
		log.Info("Closing database connection...")
		dbpool.Close()
		log.Info("Database connection closed")
	}()

	teamRepo := postgres.NewTeamRepository(dbpool)
	userRepo := postgres.NewUserRepository(dbpool)
	reviewerRepo := postgres.NewReviewerRepository(dbpool)
	prRepo := postgres.NewPullRequestRepository(dbpool, reviewerRepo)
	statsRepo := postgres.NewStatisticsRepository(dbpool)

	teamService := service.NewTeamService(teamRepo, userRepo, log)
	userService := service.NewUserService(userRepo, prRepo, reviewerRepo, log)
	pullRequestService := service.NewPullRequestService(prRepo, userRepo, reviewerRepo, log)
	statisticsService := service.NewStatisticsService(statsRepo, log)

	handlers := handler.NewHandlers(teamService, userService, pullRequestService, statisticsService, log)

	server := httpserver.NewServer(log, config, handlers)

	serverDone := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server...")
		if err := server.Run(); err != nil {
			serverDone <- err
		}
		close(serverDone)
	}()

	select {
	case sig := <-sigChan:
		log.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()

		if err := server.Shutdown(); err != nil {
			log.Error("Failed to shutdown HTTP server", zap.Error(err))
		}

		log.Info("Waiting for goroutines to finish...")

		log.Info("Application gracefully shut down")
		return nil

	case err := <-serverDone:
		if err != nil {
			log.Error("HTTP server stopped with error", zap.Error(err))
			cancel()
			return err
		}
		return nil
	}
}
