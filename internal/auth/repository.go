package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) UserRepository {
	return &pgUserRepository{
		pool: pool,
	}
}

func (r *pgUserRepository) CreateUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	const op = "auth.repository.CreateUser"

	var userID uuid.UUID

	query := "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id"
	err := r.pool.QueryRow(ctx, query, email, passwordHash).Scan(&userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return uuid.Nil, errors.New("email already 1")
			}
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (r *pgUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const op = "auth.repository.GetUserByEmail"

	var user User

	query := "SELECT id, email, password_hash, created_at FROM users WHERE email = $1"
	err := r.pool.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email, &user.PasswordHash, user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, errors.New("email not found"))
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
