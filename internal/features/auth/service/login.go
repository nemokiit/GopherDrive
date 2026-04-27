package service

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(
	ctx context.Context,
	email,
	password string,
) (user core_domain.User, accessToken string, refreshToken string, err error) {
	const op = "auth.service.Login"

	user, err = s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, core_errors.ErrInvalidCredentials)
	}
	if err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, refreshToken, err = s.generateAndSaveTokens(ctx, user.ID)
	if err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	return user, accessToken, refreshToken, nil
}
