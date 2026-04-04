package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func New(addr string, pass string) (*redis.Client, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
	})

	pingContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisCli.Ping(pingContext).Err(); err != nil {
		_ = redisCli.Close()
		return nil, fmt.Errorf("redis database ping failed: %w", err)
	}
	
	return redisCli, nil
}
