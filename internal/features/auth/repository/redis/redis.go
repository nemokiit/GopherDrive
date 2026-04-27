package redis

import (
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	redisCli *redis.Client
}

func NewRepo(redisCli *redis.Client) *Repository {
	return &Repository{
		redisCli: redisCli,
	}
}
