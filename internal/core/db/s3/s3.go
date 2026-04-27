package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func New(addr, user, password, bucketName string) (*s3.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	credentialsProvider := credentials.NewStaticCredentialsProvider(user, password, "")
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentialsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 config: %w", err)
	}

	s3Cli := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(addr)
		o.UsePathStyle = true
	})

	ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	buckets, err := s3Cli.ListBuckets(ctxPing, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to s3 (ping): %w", err)
	}

	exists := false
	for _, bucket := range buckets.Buckets {
		if aws.ToString(bucket.Name) == bucketName {
			exists = true
			break
		}
	}

	if !exists {
		_, err = s3Cli.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return s3Cli, nil
}
