package redisClient

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address  string `envconfig:"ADDRESS" required:"true"`
	Password string `envconfig:"PASSWORD" default:""`
}

func NewConfigMust() (config Config) {
	if err := envconfig.Process("REDIS", &config); err != nil {
		panic("failed to read redis environment config: " + err.Error())
	}

	return config
}
