package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucketName = "pk-positive-news" // Change this to your S3 bucket name
	objectKey  = "latest_news.json" // File that stores the latest 10 articles
	region     = "us-east-2"        // Change to your AWS region
	indexKey   = "index.html"       // File to update
)

// UploadJSONToS3 uploads JSON data to an S3 bucket
func UploadJSONToS3(ctx context.Context, data interface{}) error {
	cfg, _ := LoadAWSConfigWithRegion(ctx, region)

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
	cfg, _ := LoadAWSConfigWithRegion(ctx, region)

	s3Client := s3.NewFromConfig(cfg)
	psClient := s3.NewPresignClient(s3Client)

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

// UpdateIndexHTML replaces the pre-signed URL inside index.html and uploads the new version to S3
func UpdateIndexHTML(ctx context.Context, preSignedURL string) error {
	cfg, _ := LoadAWSConfigWithRegion(ctx, region)
	s3Client := s3.NewFromConfig(cfg)

	// Fetch the existing index.html from S3
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(indexKey),
	}

	resp, err := s3Client.GetObject(ctx, getInput)
	if err != nil {
		return fmt.Errorf("failed to fetch existing index.html: %w", err)
	}
	defer resp.Body.Close()

	// Read the file content
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read index.html content: %w", err)
	}
	htmlContent := string(bodyBytes) // Convert bytes to string

	// Replace the old pre-signed URL with the new one
	updatedHTML := strings.Replace(htmlContent, `"https://your-s3-bucket.s3.amazonaws.com/latest_news.json?...signed-url-params"`, `"`+preSignedURL+`"`, 1)

	// Upload the modified index.html back to S3
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(indexKey),
		Body:        strings.NewReader(updatedHTML),
		ContentType: aws.String("text/html"),
	}

	_, err = s3Client.PutObject(ctx, putInput)
	if err != nil {
		return fmt.Errorf("failed to upload updated index.html: %w", err)
	}

	fmt.Println("Successfully updated and re-uploaded index.html to S3.")
	return nil
}
