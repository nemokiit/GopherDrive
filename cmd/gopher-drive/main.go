package main

import (
	"GopherDrive/internal/core/config"
	"GopherDrive/internal/core/logger"
	postgres_pgx "GopherDrive/internal/core/repository/postgres/pool/pgx"
	redis_client "GopherDrive/internal/core/repository/redis/client"
	s3_client "GopherDrive/internal/core/repository/s3/client"
	transport_middleware "GopherDrive/internal/core/transport/http/middleware"
	"GopherDrive/internal/core/transport/http/server"
	auth_postgres "GopherDrive/internal/features/auth/repository/postgres"
	auth_redis "GopherDrive/internal/features/auth/repository/redis"
	auth_service "GopherDrive/internal/features/auth/service"
	auth_transport "GopherDrive/internal/features/auth/transport"
	storage_postgres "GopherDrive/internal/features/storage/repository/postgres"
	storage_s3 "GopherDrive/internal/features/storage/repository/s3"
	storage_service "GopherDrive/internal/features/storage/service"
	storage_transport "GopherDrive/internal/features/storage/transport"
	sl "GopherDrive/internal/pkg/api/logger"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	log := logger.SetupLogger(logger.NewConfigMust())

	log.Info("initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")
	log.Debug("application time zone", slog.Any("zone", time.Local))

	postgresCli, err := postgres_pgx.New(ctx, postgres_pgx.NewConfigMust())
	if err != nil {
		log.Error("failed to init postgres DB", sl.SlogErr(err))
		os.Exit(1)
	}
	defer postgresCli.Close()

	redisCli, err := redis_client.New(ctx, redis_client.NewConfigMust())
	if err != nil {
		log.Error("failed to init redis DB", sl.SlogErr(err))
		os.Exit(1)
	}
	defer func() { _ = redisCli.Close() }()

	s3Config := s3_client.NewConfigMust()
	s3Cli, err := s3_client.New(ctx, s3Config)
	if err != nil {
		log.Error("failed to init minIO DB", sl.SlogErr(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(transport_middleware.NewLogger(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	authPostgresRepository := auth_postgres.NewRepo(postgresCli)
	authRedisRepository := auth_redis.NewRepo(redisCli)
	authService := auth_service.New(authPostgresRepository, authRedisRepository, cfg.SecretToken)
	authTransport := auth_transport.New(authService, log)

	storagePostgresRepo := storage_postgres.NewRepo(postgresCli)
	storageS3Repo := storage_s3.NewRepo(s3Cli, s3Config.BucketName)
	storageService := storage_service.New(storagePostgresRepo, storageS3Repo)
	storageTransport := storage_transport.New(log, storageService)

	router.Route("/user", func(r chi.Router) {
		r.Post("/register", authTransport.Register)
		r.Post("/login", authTransport.Login)
	})

	router.Group(func(r chi.Router) {
		r.Use(transport_middleware.Authenticate(log, cfg.SecretToken))

		r.Route("/folder", func(r chi.Router) {
			r.Post("/", storageTransport.CreateFolder)
			r.Get("/", storageTransport.GetFolderContent)
			r.Get("/{id}", storageTransport.GetFolderContent)
			r.Delete("/{id}", storageTransport.DeleteFolder)
		})

		r.Route("/file", func(r chi.Router) {
			r.Post("/", storageTransport.UploadFile)
			r.Get("/{id}/download", storageTransport.DownloadFile)
			r.Delete("/{id}", storageTransport.DeleteFile)
		})
	})

	httpServer := server.NewHTTPServer(log, router, server.NewConfigMust())
	if err = httpServer.Run(ctx); err != nil {
		log.Error("HTTP server run error: ", sl.SlogErr(err))
	}
}
