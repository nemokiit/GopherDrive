package postgres

import (
	core_domain "GopherDrive/internal/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func fillFile(file *core_domain.File, row pgx.Row) error {
	err := row.Scan(
		&file.ID,
		&file.Name,
		&file.UserID,
		&file.FolderID,
		&file.S3Key,
		&file.Size,
		&file.ContentType,
		&file.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
