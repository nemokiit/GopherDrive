package logger

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env string `envconfig:"ENV" default:"PROD"`
}

func NewConfigMust() (config Config) {
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("failed to read environment for logger: %s", err)
	}

	return config
}
