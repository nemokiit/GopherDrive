package main

import (
	"GopherDrive/internal/auth/handler"
	"GopherDrive/internal/auth/repository"
	"GopherDrive/internal/auth/service"
	"GopherDrive/internal/infra/config"
	"GopherDrive/internal/infra/db/postgres"
	"GopherDrive/internal/infra/db/redis"
	"GopherDrive/internal/infra/logger"
	mwLogger "GopherDrive/internal/infra/middleware"
	sl "GopherDrive/internal/lib/api/logger"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")

	postgresDB, err := postgres.New(cfg.PostgresURL)
	if err != nil {
		log.Error("failed to init postgres DB", sl.SlogErr(err))
		os.Exit(1)
	}
	defer postgresDB.Close()

	redisCli, err := redis.New(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Error("failed to init redis DB", sl.SlogErr(err))
		os.Exit(1)
	}
	defer func() { _ = redisCli.Close() }()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.NewLogger(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	authPostgresRepository := repository.NewPostgresRepository(postgresDB)
	authService := service.NewService(authPostgresRepository)
	authHandlers := handler.NewHandler(authService, log)

	router.Route("/user", func(r chi.Router) {
		r.Post("/register", authHandlers.Register)
		r.Post("/login", authHandlers.Login)
	})

	log.Info("server starting", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Error("failed to start server", sl.SlogErr(err))
		}
	}()

	log.Info("server started")

	<-done
	log.Info("server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.SlogErr(err))
		return
	}

	log.Info("server stopped")
}
