package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	Address     string
	SecretToken string
	TimeZone    *time.Location
}

func NewConfigMust() (config *Config) {
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		log.Fatalf("SERVER_ADDRESS is not set in environment")
	}

	secretToken := os.Getenv("SECRET_TOKEN")
	if secretToken == "" {
		log.Fatalf("SECRET_TOKEN is not set in environment")
	}

	tz := os.Getenv("TIME_ZONE")

	zone, err := time.LoadLocation(tz)
	if err != nil {
		log.Fatalf("failed to load time zone: %s: %s", tz, err)
	}

	return &Config{
		Address:     address,
		SecretToken: secretToken,
		TimeZone:    zone,
	}
}
