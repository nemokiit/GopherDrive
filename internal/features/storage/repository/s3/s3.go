package s3

import (
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Repository struct {
	s3Cli      *s3.Client
	uploader   *transfermanager.Client
	bucketName string
}

func NewRepo(s3Cli *s3.Client, bucketName string) *Repository {
	uploader := transfermanager.New(s3Cli, func(o *transfermanager.Options) {
		o.PartSizeBytes = manager.DefaultDownloadPartSize
		o.Concurrency = manager.DefaultUploadConcurrency
	})

	return &Repository{
		s3Cli:      s3Cli,
		uploader:   uploader,
		bucketName: bucketName,
	}
}
