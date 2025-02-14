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

// BuildPlainMessage generates the email content with the top articles and a pre-signed S3 URL
// BuildPlainMessage generates the email content with a top 10 news list and a link to the website
func BuildPlainMessage(topArticles []ArticleWithContent, preSignedURL string) string {
	plainMessage := "Hello,\n\n"
	plainMessage += "Here are your top 10 positively ranked articles for today:\n\n"

	plainMessage += "Check out the latest positive news articles on our website 🌟: http://bit.ly/3CNTB7C\n\n"

	for i, art := range topArticles {
		plainMessage += fmt.Sprintf("%d. %s\n%s\n\n", i+1, art.Title, art.URL)
	}

	plainMessage += "\nHave a wonderful day!\n"

	return plainMessage
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
