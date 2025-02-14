// email.go
package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// buildPlainMessage builds the plain text email message.
func BuildPlainMessage(topArticles []ArticleWithContent, rankedArticles []RankedArticle) string {
	message := "Hello,\n\nHere are your top positively-ranked news articles:\n\n"
	for _, art := range topArticles {
		message += fmt.Sprintf("- %s %s\n\n", art.Title, art.URL)
	}
	message += "\nFull Ranking Details:\n\n"
	for _, ra := range rankedArticles {
		message += fmt.Sprintf("%d. %s %s - Category: %s\n\n", ra.Rank, ra.Title, ra.URL, ra.Category)
	}
	message += "\nHave a great day!\n"
	return message
}

// sendEmailViaSNS sends a plain text email via SNS by ensuring each link is on its own line.
func SendEmailViaSNS(ctx context.Context, topicARN, subject, message string) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-2"))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	snsClient := sns.NewFromConfig(cfg)
	lines := strings.Split(message, "\n")
	for i, line := range lines {
		if strings.Contains(line, "http") {
			lines[i] = line + "\n"
		}
	}
	finalMessage := strings.Join(lines, "\n")
	input := &sns.PublishInput{
		Message:  aws.String(finalMessage),
		TopicArn: aws.String(topicARN),
		Subject:  aws.String(subject),
	}
	_, err = snsClient.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}
