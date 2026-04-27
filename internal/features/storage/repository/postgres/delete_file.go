package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) DeleteFile(ctx context.Context, userID, fileID uuid.UUID) (file core_domain.File, err error) {
	const op = "storage.repository.DeleteFile"

	query := "DELETE FROM files WHERE user_id = $1 AND id = $2 RETURNING *"
	row := r.pool.QueryRow(ctx, query, userID, fileID)

	err = fillFile(&file, row)
	if errors.Is(err, pgx.ErrNoRows) {
		return file, fmt.Errorf("%s: %w", op, core_errors.ErrFileNotFound)
	}
	if err != nil {
		return file, fmt.Errorf("%s: %w", op, err)
	}

	return file, nil
}
