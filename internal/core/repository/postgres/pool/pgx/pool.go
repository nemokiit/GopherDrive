package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, config Config) (pool *pgxpool.Pool, err error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err = pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err = pool.Ping(ctxPing); err != nil {
		return nil, fmt.Errorf("postgres database ping failed: %w", err)
	}

	return pool, nil
}
