package pgx

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     string `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	Database string `envconfig:"DATABASE" required:"true"`
}

func NewConfigMust() (config Config) {
	if err := envconfig.Process("POSTGRES", &config); err != nil {
		panic("failed to read postgres environment config: " + err.Error())
	}

	return config
}
