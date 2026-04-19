package repository

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Repository struct {
	s3Cli      *s3.Client
	bucketName string
}

func NewS3Repo(s3Cli *s3.Client, bucketName string) *S3Repository {
	return &S3Repository{
		s3Cli:      s3Cli,
		bucketName: bucketName,
	}
}

func (r *S3Repository) UploadFile(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	const op = "storage.repository.UploadFile"

	_, err := r.s3Cli.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(key),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *S3Repository) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	const op = "storage.repository.DownloadFile"

	getObj, err := r.s3Cli.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	return getObj.Body, nil
}

func (r *S3Repository) DeleteFile(ctx context.Context, key string) error {
	const op = "storage.repository.DeleteFile"

	_, err := r.s3Cli.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
