package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	poolPG *pgxpool.Pool
}

func NewRepo(poolPG *pgxpool.Pool) *Repository {
	return &Repository{
		poolPG: poolPG,
	}
}
