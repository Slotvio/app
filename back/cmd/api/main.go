package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/aleksandrboro/slotvio/internal/handler"
	postgres "github.com/aleksandrboro/slotvio/internal/repository"
	"github.com/aleksandrboro/slotvio/internal/router"
	"github.com/aleksandrboro/slotvio/internal/server"
	"github.com/aleksandrboro/slotvio/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	envDev     = "dev"
	envStaging = "staging"
	envProd    = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.App.Env)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := createPgxPool(ctx, &cfg.DB)
	if err != nil {
		panic(err)
	}

	masterRepo, err := postgres.NewMasterRepo(pool)
	if err != nil {
		panic(err)
	}

	masterService, err := service.NewMasterService(log, masterRepo)
	if err != nil {
		panic(err)
	}
	masterHandler := handler.NewMasterHandler(log, masterService)
	router, err := router.NewRouter(masterHandler)
	if err != nil {
		panic(err)
	}

	server, err := server.NewServer(log, cfg, router)
	if err != nil {
		panic(err)
	}

	go server.Start()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envStaging:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func createPgxPool(ctx context.Context, dbCfg *config.DBCfg) (*pgxpool.Pool, error) {
	if dbCfg == nil {
		return nil, errors.New("empty db config")
	}

	connURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbCfg.User, dbCfg.Password),
		Host:   net.JoinHostPort(dbCfg.Host, dbCfg.Port),
		Path:   "/" + dbCfg.Name,
	}

	pool, err := pgxpool.New(ctx, connURL.String())
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	return pool, nil
}
