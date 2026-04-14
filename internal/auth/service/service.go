package service

import (
	"GopherDrive/internal/auth"
	"GopherDrive/internal/lib/tokens"
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

func (s *Service) Register(ctx context.Context, email, password string) (uuid.UUID, string, string, error) {
	const op = "auth.service.Register"

	if err := isValidEmail(email); err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.users.CreateUser(ctx, email, string(passwordHash))
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, refreshToken, err := s.generateAndSaveTokens(ctx, id)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	return id, accessToken, refreshToken, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (uuid.UUID, string, string, error) {
	const op = "auth.service.Login"

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, auth.ErrInvalidCredentials)
	}
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, refreshToken, err := s.generateAndSaveTokens(ctx, user.ID)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, accessToken, refreshToken, nil
}

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

func (s *Service) generateAndSaveTokens(ctx context.Context, id uuid.UUID) (string, string, error) {
	accessToken, err := tokens.GenerateAccessToken(s.secretToken, id, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := tokens.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	err = s.sessions.SaveRefreshToken(ctx, refreshToken, id, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
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
