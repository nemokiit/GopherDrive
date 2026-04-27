package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (r *Repository) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
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
