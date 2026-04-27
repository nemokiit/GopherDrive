package service

import (
	core_domain "GopherDrive/internal/core/domain"
	"GopherDrive/internal/pkg/tokens"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepo interface {
	SaveUser(ctx context.Context, user core_domain.User) error
	GetUserByEmail(ctx context.Context, email string) (core_domain.User, error)
}

type SessionRepo interface {
	SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type Service struct {
	users       UserRepo
	sessions    SessionRepo
	secretToken string
}

func New(u UserRepo, s SessionRepo, secretToken string) *Service {
	return &Service{
		users:       u,
		sessions:    s,
		secretToken: secretToken,
	}
}

func (s *Service) generateAndSaveTokens(
	ctx context.Context,
	id uuid.UUID,
) (accessToken string, refreshToken string, err error) {
	accessToken, err = tokens.GenerateAccessToken(s.secretToken, id, 15*time.Minute)
	if err != nil {
		return accessToken, refreshToken, err
	}

	refreshToken, err = tokens.GenerateRefreshToken()
	if err != nil {
		return accessToken, refreshToken, err
	}

	err = s.sessions.SaveRefreshToken(ctx, refreshToken, id, 24*time.Hour)
	if err != nil {
		return accessToken, refreshToken, err
	}

	return accessToken, refreshToken, nil
}
