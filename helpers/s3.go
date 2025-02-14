package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/presign"
)

const (
	bucketName = "pk-positive-news" // Change this to your S3 bucket name
	objectKey  = "latest_news.json" // File that stores the latest 10 articles
	region     = "us-east-2"        // Change to your AWS region
)

// UploadJSONToS3 uploads JSON data to an S3 bucket
func UploadJSONToS3(ctx context.Context, data interface{}) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        strings.NewReader(string(jsonData)),
		ContentType: aws.String("application/json"),
	}

	_, err = s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload JSON to S3: %w", err)
	}

	fmt.Println("Successfully uploaded latest news to S3:", objectKey)
	return nil
}

// GeneratePreSignedURL creates a temporary S3 URL for latest_news.json
func GeneratePreSignedURL(ctx context.Context) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	psClient := presign.NewPresignClient(s3Client)

	// Create a pre-signed URL that expires in 24 hours
	req, err := psClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	return req.URL, nil
}
