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

// DefaultBucketName is the bucket ensured on startup and used throughout the
// app. Kept as a constant so it can be swapped later (e.g. per-env or per-org).
const DefaultBucketName = "test-bucket"

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

	ensureDefaultBucket(ctx)
}

// ensureDefaultBucket makes sure DefaultBucketName exists so downstream code
// paths (uploads, presigned URLs) don't fail on a fresh MinIO volume. Any
// failure is logged but does not abort startup.
func ensureDefaultBucket(ctx context.Context) {
	exists, err := S3BasicsBucket.BucketExists(ctx, DefaultBucketName)
	if err != nil {
		fmt.Printf("Failed to check if bucket %q exists: %v\n", DefaultBucketName, err)
		return
	}
	if exists {
		fmt.Printf("Bucket %q already exists, skipping creation.\n", DefaultBucketName)
		return
	}
	region := os.Getenv("AWS_REGION")
	if err := S3BasicsBucket.CreateBucket(ctx, DefaultBucketName, region); err != nil {
		fmt.Printf("Failed to create bucket %q in region %q: %v\n", DefaultBucketName, region, err)
		return
	}
	fmt.Printf("Created bucket %q in region %q.\n", DefaultBucketName, region)
}
