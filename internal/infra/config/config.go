package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env"`
	SecretToken string `yaml:"secret_token"`

	HTTPServer     `yaml:"http_server"`
	PostgresConfig `yaml:"postgres_config"`
	RedisConfig    `yaml:"redis_config"`
}

type HTTPServer struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type PostgresConfig struct {
	PostgresURL string `yaml:"pg_url" env-required:"true"`
}

type RedisConfig struct {
	RedisAddr string `yaml:"redis_addr" env-required:"true"`
	RedisPass string `yaml:"redis_pass"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH env variable is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &config
}
