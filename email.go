// email.go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// sendEmailViaSNS sends a plain text email via SNS by ensuring each link is on its own line.
func sendEmailViaSNS(ctx context.Context, topicARN, subject, message string) error {
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
