// common.go
package helpers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Shared constants
const (
	NewsAPIURL           = "https://newsapi.org/v2/everything"
	TableName            = "PositiveArticles"
	SnsTopicARNHardcoded = "arn:aws:sns:us-east-2:969666470832:positive_news"
)

// Shared type definitions
type NewsResponse struct {
	Articles []Article `json:"articles"`
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	ImageURL    string `json:"urlToImage"`
}

type ArticleWithContent struct {
	Title    string
	URL      string
	Excerpt  string
	ImageURL string
}

type RankedArticle struct {
	Rank     int    `json:"rank"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return cfg, fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	return cfg, nil
}

// LoadAWSConfigWithRegion loads the AWS configuration for a specified region.
func LoadAWSConfigWithRegion(ctx context.Context, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS configuration for region %s: %w", region, err)
	}
	return cfg, nil
}
