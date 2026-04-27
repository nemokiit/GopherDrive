package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (r *Repository) SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	const op = "auth.repository.SaveRefreshToken"

	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.redisCli.Set(ctx, key, userID.String(), ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
