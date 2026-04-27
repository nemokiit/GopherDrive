package redis

import (
	"context"
	"fmt"
)

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	const op = "auth.repository.DeleteToken"

	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.redisCli.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
