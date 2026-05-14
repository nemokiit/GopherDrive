package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	log    *slog.Logger
	router chi.Router
	cfg    Config
}

func NewHTTPServer(log *slog.Logger, router chi.Router, cfg Config) *HTTPServer {
	return &HTTPServer{
		log:    log,
		router: router,
		cfg:    cfg,
	}
}

func (h *HTTPServer) Run(ctx context.Context) error {
	h.log.Info("server starting", slog.String("address", h.cfg.Address))

	server := &http.Server{
		Addr:         h.cfg.Address,
		Handler:      h.router,
		ReadTimeout:  h.cfg.Timeout,
		WriteTimeout: h.cfg.Timeout,
		IdleTimeout:  h.cfg.IdleTimeout,
	}

	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	h.log.Info("server started")

	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	case <-ctx.Done():
		h.log.Info("server stopping")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), h.cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("failed to stop server: %w", err)
		}
	}

	h.log.Info("server stopped")

	return nil
}
