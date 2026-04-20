package repository

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func (r *S3Repository) DeleteFiles(ctx context.Context, keys []string) error {
	const op = "storage.repository.DeleteFiles"

	if len(keys) == 0 {
		return nil
	}

	var errs error
	for i := 0; i < len(keys); i += 1000 {
		end := i + 1000

		if len(keys) < end {
			end = len(keys)
		}

		batch := keys[i:end]

		objects := make([]types.ObjectIdentifier, 0, len(batch))
		for _, key := range batch {
			objects = append(objects, types.ObjectIdentifier{Key: aws.String(key)})
		}

		input := s3.DeleteObjectsInput{
			Bucket: aws.String(r.bucketName),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		}

		output, err := r.s3Cli.DeleteObjects(ctx, &input)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if len(output.Errors) > 0 {
			for _, delErr := range output.Errors {
				errs = errors.Join(errs, fmt.Errorf("failed to delete key %s: %s", aws.ToString(delErr.Key),
					aws.ToString(delErr.Message)))
			}
		}
	}

	if errs != nil {
		return fmt.Errorf("%s: %w", op, errs)
	}

	return nil
}
