package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("email not found")
	ErrUserAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTokenNotFound = errors.New("token not found")
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
