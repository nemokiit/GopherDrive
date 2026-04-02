package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err = pool.Ping(ctxPing); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return pool, nil
}
