package core_domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func CreateUser(email, passwordHash string) User {
	var (
		id        = uuid.New()
		createdAt = time.Now()
	)

	return User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}
}
