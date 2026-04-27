package transport

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type Service interface {
	Register(ctx context.Context, email string, password string) (core_domain.User, string, string, error)
	Login(ctx context.Context, email string, password string) (core_domain.User, string, string, error)
	Refresh(ctx context.Context, cookie string) (uuid.UUID, string, string, error)
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func New(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}
