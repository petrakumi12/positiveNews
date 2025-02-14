package helpers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// NewSNSClient creates and returns an SNS client for the specified region.
func NewSNSClient(ctx context.Context, region string) (*sns.Client, error) {
	cfg, err := LoadAWSConfigWithRegion(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return sns.NewFromConfig(cfg), nil
}

// SubscribeUser subscribes the given email to the SNS topic.
// It uses the SNS client to call the Subscribe API.
func SubscribeUser(ctx context.Context, topicARN string, email string) error {
	client, err := NewSNSClient(ctx, "us-east-2")
	if err != nil {
		return fmt.Errorf("failed to create SNS client: %w", err)
	}

	_, err = client.Subscribe(ctx, &sns.SubscribeInput{
		Protocol: aws.String("email"),
		Endpoint: aws.String(email),
		TopicArn: aws.String(topicARN),
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe %s: %w", email, err)
	}
	return nil
}

// UnsubscribeUser removes an email subscription from the SNS topic.
// It returns a message and an error.
func UnsubscribeUser(ctx context.Context, topicARN string, email string) (string, error) {
	client, err := NewSNSClient(ctx, "us-east-2")
	if err != nil {
		return "", fmt.Errorf("failed to create SNS client: %w", err)
	}

	// List subscriptions for the topic.
	input := &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(topicARN),
	}

	result, err := client.ListSubscriptionsByTopic(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var subscriptionArn string
	for _, sub := range result.Subscriptions {
		if sub.Endpoint != nil && *sub.Endpoint == email {
			subscriptionArn = *sub.SubscriptionArn
			break
		}
	}

	// If no subscription is found, return a success message indicating nothing to do.
	if subscriptionArn == "" {
		return fmt.Sprintf("No active subscription found for %s", email), nil
	}

	// If subscription is still pending confirmation, report that as a successful response.
	if subscriptionArn == "PendingConfirmation" {
		return fmt.Sprintf("Subscription for %s is still pending confirmation; no action taken", email), nil
	}

	// Unsubscribe the user.
	_, err = client.Unsubscribe(ctx, &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(subscriptionArn),
	})
	if err != nil {
		return "", fmt.Errorf("failed to unsubscribe %s: %w", email, err)
	}

	return fmt.Sprintf("Successfully unsubscribed %s", email), nil
}

// SendEmailViaSNS sends a plain text email via SNS with the given subject and message.
func SendEmail(ctx context.Context, topicARN, subject, message string) error {
	client, err := NewSNSClient(ctx, "us-east-2")
	if err != nil {
		return err
	}

	input := &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Subject:  aws.String(subject),
		Message:  aws.String(message),
	}

	_, err = client.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish SNS message: %w", err)
	}

	return nil
}
