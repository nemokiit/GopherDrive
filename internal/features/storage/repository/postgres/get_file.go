package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	"GopherDrive/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetFileByID(ctx context.Context, userID, fileID uuid.UUID) (file core_domain.File, err error) {
	const op = "storage.repository.GetFileByID"

	query := "SELECT * FROM files WHERE user_id = $1 AND id = $2"
	row := r.pool.QueryRow(ctx, query, userID, fileID)

	err = fillFile(&file, row)
	if errors.Is(err, pgx.ErrNoRows) {
		return file, fmt.Errorf("%s: %w", op, pool.ErrFileNotFound)
	}
	if err != nil {
		return file, fmt.Errorf("%s: %w", op, err)
	}

	return file, nil
}
