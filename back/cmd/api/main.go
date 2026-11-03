package main

import (
	"log/slog"
	"os"

	"github.com/aleksandrboro/slotvio/internal/config"
)

const (
	envDev     = "dev"
	envStaging = "staging"
	envProd    = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.App.Env)
	log.Info("config", slog.Any("config", cfg))
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
