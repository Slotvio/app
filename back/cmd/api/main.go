package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/aleksandrboro/slotvio/internal/server"
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

	server, err := server.NewServer(log, cfg, nil)
	if err != nil {
		panic(err)
	}

	server.Start()

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
