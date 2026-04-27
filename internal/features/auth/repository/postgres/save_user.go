package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) SaveUser(ctx context.Context, user core_domain.User) error {
	const op = "auth.repository.CreateUser"

	query := "INSERT INTO users (id, email, password_hash, created_at) VALUES ($1, $2)"
	_, err := r.poolPG.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, core_errors.ErrUserNotFound)
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
