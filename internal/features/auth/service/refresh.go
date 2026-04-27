package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) Refresh(ctx context.Context, cookie string) (uuid.UUID, string, string, error) {
	const op = "auth.service.Refresh"

	id, err := s.sessions.GetUserIDByToken(ctx, cookie)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = s.sessions.DeleteRefreshToken(ctx, cookie)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, refreshToken, err := s.generateAndSaveTokens(ctx, id)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	return id, accessToken, refreshToken, nil
}
