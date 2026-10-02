package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	keyPrefix    = "assets/"
	cacheControl = "public, max-age=31536000, immutable"
)

type S3Config struct {
	Endpoint        string // https://<account>.r2.cloudflarestorage.com
	Region          string // "auto" for R2
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	AssetBaseURL    string // public URL of the assets/ prefix, e.g. https://cdn.example.com/assets
}

type S3 struct {
	client       *s3.Client
	bucket       string
	assetBaseURL string
}

func NewS3(cfg S3Config) (*S3, error) {

	if cfg.Endpoint == "" {
		return nil, errors.New("storage: endpoint is required")
	}

	if cfg.Bucket == "" {
		return nil, errors.New("storage: bucket is required")
	}

	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, errors.New("storage: access key id and secret access key are required")
	}
	if cfg.AssetBaseURL == "" {
		return nil, errors.New("storage: asset base url is required")
	}

	region := cfg.Region
	if region == "" {
		region = "auto"
	}

	client := s3.New(s3.Options{
		Region:                     region,
		BaseEndpoint:               aws.String(cfg.Endpoint),
		Credentials:                credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		UsePathStyle:               true,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})

	return &S3{
		client:       client,
		bucket:       cfg.Bucket,
		assetBaseURL: strings.TrimRight(cfg.AssetBaseURL, "/"),
	}, nil
}

func (s *S3) Put(ctx context.Context, hash, contentType string, size int64, body io.ReadSeeker) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(keyPrefix + hash),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
		CacheControl:  aws.String(cacheControl),
	})
	if err != nil {
		return fmt.Errorf("storage: put %s: %w", hash, err)
	}
	return nil
}

func (s *S3) PublicURL(hash string) string {
	return s.assetBaseURL + "/" + hash
}
