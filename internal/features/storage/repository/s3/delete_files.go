package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (r *Repository) DeleteFiles(ctx context.Context, keys []string) error {
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
