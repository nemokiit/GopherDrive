package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (user core_domain.User, err error) {
	const op = "auth.repository.GetUserByEmail"

	query := "SELECT id, email, password_hash, created_at FROM users WHERE email = $1"
	err = r.poolPG.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("%s: %w", op, core_errors.ErrUserNotFound)
		}

		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
