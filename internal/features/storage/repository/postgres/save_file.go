package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) SaveFile(ctx context.Context, file core_domain.File) error {
	const op = "storage.repository.CreateFile"

	query := "INSERT INTO files (name, user_id, folder_id, s3_key, size, content_type) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := r.pool.Exec(ctx, query, file.Name, file.UserID, file.FolderID, file.S3Key, file.Size, file.ContentType)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%s: %w", op, core_errors.ErrFileAlreadyExists)
		case "23503":
			switch pgErr.ConstraintName {
			case "files_user_id_fkey":
				return fmt.Errorf("%s: %w", op, core_errors.ErrUserNotFound)
			case "files_folder_id_fkey":
				return fmt.Errorf("%s: %w", op, core_errors.ErrFolderNotFound)
			default:
				return fmt.Errorf("%s: foreign key violation for %s", op, pgErr.ConstraintName)
			}
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
