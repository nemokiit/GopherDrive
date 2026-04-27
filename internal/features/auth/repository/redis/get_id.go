package redis

import (
	core_errors "GopherDrive/internal/core/errors"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Repository) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	const op = "auth.repository.GetUserIDByToken"

	var strID string

	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.redisCli.Get(ctx, key).Scan(&strID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, core_errors.ErrTokenNotFound
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := uuid.Parse(strID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
