package service

import (
	"GopherDrive/internal/auth"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*auth.User, error)
}

type sessionRepo interface {
	SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type Service struct {
	users    UserRepo
	sessions SessionRepo
}

func New(u UserRepo, s SessionRepo) *Service {
	return &Service{
		users:    u,
		sessions: s,
	}
}

func (s *Service) Register(ctx context.Context, email string, password string) (uuid.UUID, error) {
	const op = "auth.service.Register"

	if err := isValidEmail(email); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.users.CreateUser(ctx, email, string(passwordHash))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (uuid.UUID, error) {
	const op = "auth.service.Login"

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return uuid.Nil, fmt.Errorf("%s: %w", op, auth.ErrInvalidCredentials)
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
