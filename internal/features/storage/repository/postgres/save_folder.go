package postgres

import (
	core_domain "GopherDrive/internal/core/domain"
	"GopherDrive/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) SaveFolder(ctx context.Context, folder core_domain.Folder) error {
	const op = "storage.repository.CreateFolder"

	query := "INSERT INTO folders (name, user_id, parent_folder_id) VALUES ($1, $2, $3)"
	_, err := r.pool.Exec(ctx, query, folder.Name, folder.UserID, folder.ParentFolderID)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%s: %w", op, pool.ErrFolderAlreadyExists)
		case "23503":
			switch pgErr.ConstraintName {
			case "folders_user_id_fkey":
				return fmt.Errorf("%s: %w", op, pool.ErrUserNotFound)
			case "folders_parent_folder_id_fkey":
				return fmt.Errorf("%s: %w", op, pool.ErrParentFolderNotFound)
			default:
				return fmt.Errorf("%s: foreign key violation for %s", op, pgErr.ConstraintName)
			}
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
