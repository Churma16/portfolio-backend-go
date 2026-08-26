package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Config holds configuration for Cloudflare R2 connection
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

// LoadR2ConfigFromEnv reads R2 configuration from environment variables
func LoadR2ConfigFromEnv() R2Config {
	return R2Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("R2_BUCKET_NAME"),
		PublicURL:       os.Getenv("R2_PUBLIC_URL"),
	}
}

// Validate checks if all required fields for Cloudflare R2 are present
func (c *R2Config) Validate() error {
	if c.AccountID == "" || c.AccessKeyID == "" || c.SecretAccessKey == "" || c.BucketName == "" {
		return fmt.Errorf("missing required R2 configuration (AccountID, AccessKeyID, SecretAccessKey, BucketName)")
	}
	return nil
}

// EndpointURL returns the full S3-compatible API endpoint for Cloudflare R2
func (c *R2Config) EndpointURL() string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.AccountID)
}

// NewS3Client creates and configures an AWS S3 client specifically for Cloudflare R2
func NewS3Client(ctx context.Context, r2Config R2Config) (*s3.Client, error) {
	if err := r2Config.Validate(); err != nil {
		return nil, err
	}

	r2Endpoint := r2Config.EndpointURL()

	customEndpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               r2Endpoint,
			SigningRegion:     "auto",
			HostnameImmutable: true,
		}, nil
	})

	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(customEndpointResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			r2Config.AccessKeyID,
			r2Config.SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2/S3 config: %w", err)
	}

	return s3.NewFromConfig(awsConfig), nil
}
