package storage

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsconfig "github.com/aws/smithy-go"
)

var S3BasicsBucket BucketBasics

func ConnectObjectStorage() {
	ctx := context.Background()

	s3Client := s3.NewFromConfig(aws.Config{
		Region: os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("MINIO_ROOT_USER"),
			os.Getenv("MINIO_ROOT_PASSWORD"),
			"",
		),
		BaseEndpoint: aws.String(os.Getenv("MINIO_URL")),
	}, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	presignEndpoint := os.Getenv("MINIO_PUBLIC_URL")
	if presignEndpoint == "" {
		presignEndpoint = os.Getenv("MINIO_URL")
	}

	// Use a separate client pointed at the public URL (if set) so presigned URLs
	// are signed with a host that external services (e.g. Anthropic) can reach.
	publicS3Client := s3.NewFromConfig(aws.Config{
		Region: os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("MINIO_ROOT_USER"),
			os.Getenv("MINIO_ROOT_PASSWORD"),
			"",
		),
		BaseEndpoint: aws.String(presignEndpoint),
	}, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	s3PresignClient := s3.NewPresignClient(publicS3Client)

	S3BasicsBucket = BucketBasics{
		S3Client:      s3Client,
		PresignClient: s3PresignClient,
	}

	count := 10
	result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})

	if err != nil {
		var ae awsconfig.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "AccessDenied" {
			fmt.Println("You don't have permission to list buckets for this account.")
		} else {
			fmt.Printf("Couldn't list buckets for your account. Here's why: %v\n", err)
		}
		return
	}

	if len(result.Buckets) == 0 {
		fmt.Println("You don't have any buckets!")
	} else {
		if count > len(result.Buckets) {
			count = len(result.Buckets)
		}
		for _, bucket := range result.Buckets[:count] {
			fmt.Printf("\t%v\n", *bucket.Name)
		}

		fmt.Printf("Listed %v buckets successfully\n", count)
	}
}