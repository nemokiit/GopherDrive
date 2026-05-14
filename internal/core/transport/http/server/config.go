package server

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address         string        `envconfig:"ADDRESS" required:"true"`
	Timeout         time.Duration `envconfig:"TIMEOUT" default:"4s"`
	IdleTimeout     time.Duration `envconfig:"IDLE_TIMEOUT" default:"30s"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

func NewConfigMust() (config Config) {
	if err := envconfig.Process("SERVER", &config); err != nil {
		log.Fatalf("failed to read server environment config: %s", err)
	}

	return config
}
