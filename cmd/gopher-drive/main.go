package main

import (
	"GopherDrive/internal/infra/config"
	"GopherDrive/internal/infra/logger"
	"log/slog"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")
}
