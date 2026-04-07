package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, email string, password string) (uuid.UUID, error)
	Login(ctx context.Context, email string, password string) (uuid.UUID, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Register(ctx context.Context, email string, password string) (uuid.UUID, error) {
	const op = "auth.service.Register"

	if err := isValidEmail(email); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.repo.CreateUser(ctx, email, string(passwordHash))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *service) Login(ctx context.Context, email string, password string) (uuid.UUID, error) {
	const op = "auth.service.Login"

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return uuid.Nil, fmt.Errorf("%s: %w", op, errors.New("invalid credentials"))
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, nil
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
