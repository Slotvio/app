package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/aleksandrboro/slotvio/internal/handler"
	"github.com/aleksandrboro/slotvio/internal/router"
	"github.com/aleksandrboro/slotvio/internal/server"
	"github.com/aleksandrboro/slotvio/internal/service"
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

	masterService := service.NewMasterService(log, nil)
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
