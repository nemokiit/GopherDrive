package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
)

func (r *Repository) UploadFile(ctx context.Context, key string, reader io.Reader, contentType string) error {
	const op = "storage.repository.UploadFile"

	_, err := r.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
