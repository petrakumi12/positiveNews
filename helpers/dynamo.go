// dynamo.go
package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// getRecentArticleURLs retrieves URLs of articles stored in DynamoDB within the last month.
func GetRecentArticleURLs(ctx context.Context) (map[string]bool, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	ddbClient := ddb.NewFromConfig(cfg)
	oneMonthAgo := time.Now().AddDate(0, -1, 0).Format(time.RFC3339)
	input := &ddb.ScanInput{
		TableName:        aws.String(TableName),
		FilterExpression: aws.String("StoredAt >= :date"),
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
			":date": &ddbTypes.AttributeValueMemberS{Value: oneMonthAgo},
		},
	}
	result, err := ddbClient.Scan(ctx, input)
	if err != nil {
		return nil, err
	}
	recent := make(map[string]bool)
	for _, item := range result.Items {
		if urlAttr, ok := item["url"].(*ddbTypes.AttributeValueMemberS); ok {
			recent[urlAttr.Value] = true
		}
	}
	return recent, nil
}

// storeArticles saves the selected articles to DynamoDB.
func StoreArticles(ctx context.Context, articles []ArticleWithContent) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}
	ddbClient := ddb.NewFromConfig(cfg)
	for _, art := range articles {
		item := map[string]ddbTypes.AttributeValue{
			"url":      &ddbTypes.AttributeValueMemberS{Value: art.URL},
			"Title":    &ddbTypes.AttributeValueMemberS{Value: art.Title},
			"Excerpt":  &ddbTypes.AttributeValueMemberS{Value: art.Excerpt},
			"StoredAt": &ddbTypes.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		}
		input := &ddb.PutItemInput{
			TableName: aws.String(TableName),
			Item:      item,
		}
		_, err := ddbClient.PutItem(ctx, input)
		if err != nil {
			fmt.Printf("Failed to store article '%s': %v\n", art.Title, err)
		} else {
			fmt.Printf("Stored article: %s\n", art.Title)
		}
	}
	return nil
}
