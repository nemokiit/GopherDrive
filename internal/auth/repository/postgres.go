package repository

import (
	"GopherDrive/internal/auth"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepository struct {
	poolPG *pgxpool.Pool
}

func NewPostgresRepo(poolPG *pgxpool.Pool) *PGRepository {
	return &PGRepository{
		poolPG: poolPG,
	}
}

func (r *PGRepository) CreateUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	const op = "auth.repository.CreateUser"

	var userID uuid.UUID

	query := "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id"
	err := r.poolPG.QueryRow(ctx, query, email, passwordHash).Scan(&userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return uuid.Nil, fmt.Errorf("%s: %w", op, auth.ErrUserNotFound)
			}
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (r *PGRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	const op = "auth.repository.GetUserByEmail"

	var user auth.User

	query := "SELECT id, email, password_hash, created_at FROM users WHERE email = $1"
	err := r.poolPG.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, auth.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
