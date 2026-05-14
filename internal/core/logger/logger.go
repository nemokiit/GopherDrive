package logger

import (
	"log/slog"
	"os"
	"strings"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(config Config) (log *slog.Logger) {
	switch strings.ToLower(config.Env) {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	log = log.With(slog.String("env", strings.ToLower(config.Env)))

	return log
}
