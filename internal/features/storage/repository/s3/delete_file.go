package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (r *Repository) DeleteFile(ctx context.Context, key string) error {
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
