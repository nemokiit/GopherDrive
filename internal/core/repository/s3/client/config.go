package s3Client

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address      string `envconfig:"ADDRESS" required:"true"`
	RootUser     string `envconfig:"ROOT_USER" required:"true"`
	RootPassword string `envconfig:"ROOT_PASSWORD" required:"true"`
	BucketName   string `envconfig:"BUCKET_NAME" required:"true"`
	Region       string `envconfig:"REGION" default:"us-east-1"`
}

func NewConfigMust() (config Config) {
	if err := envconfig.Process("MINIO", &config); err != nil {
		panic("failed to read redis environment config: " + err.Error())
	}

	return config
}
