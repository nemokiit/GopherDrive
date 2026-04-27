package service

import (
	core_domain "GopherDrive/internal/core/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(
	ctx context.Context, email,
	password string,
) (user core_domain.User, accessToken string, refreshToken string, err error) {
	const op = "auth.service.Register"

	if err := isValidEmail(email); err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	user = core_domain.CreateUser(email, string(passwordHash))

	err = s.users.SaveUser(ctx, user)
	if err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, refreshToken, err = s.generateAndSaveTokens(ctx, user.ID)
	if err != nil {
		return user, accessToken, refreshToken, fmt.Errorf("%s: %w", op, err)
	}

	return user, accessToken, refreshToken, nil
}

func isValidEmail(email string) error {
	if len(email) < 5 {
		return errors.New("email is too short")
	}
	if len(email) > 255 {
		return errors.New("email is too long")
	}

	if strings.HasPrefix(email, "_") {
		return errors.New("email cannot start with underscore")
	}

	for _, c := range email {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != '@' && c != '.' {
			return errors.New("email contains invalid characters")
		}
	}

	return nil
}
