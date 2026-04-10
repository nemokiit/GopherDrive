package repository

import (
	"GopherDrive/internal/auth"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	redisCli *redis.Client
}

func NewRedisRepo(redisCli *redis.Client) *RedisRepository {
	return &RedisRepository{
		redisCli: redisCli,
	}
}

func (r *RedisRepository) SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	const op = "auth.repository.SaveRefreshToken"

	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.redisCli.Set(ctx, key, userID.String(), ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RedisRepository) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	const op = "auth.repository.GetUserIDByToken"

	var strID string

	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.redisCli.Get(ctx, key).Scan(&strID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, auth.ErrTokenNotFound
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := uuid.Parse(strID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *RedisRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	const op = "auth.repository.DeleteToken"

	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.redisCli.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
