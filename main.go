package main

import (
	"context"
	"encoding/json"
	"fmt"
	"positive-news/helpers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	openai "github.com/sashabaranov/go-openai"
)

// SubscriptionRequest represents a subscription/unsubscription event.
type SubscriptionRequest struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Unsubscribe bool   `json:"unsubscribe"`
}

// handleRequest dispatches based on the event content and returns a response with CORS headers.
func handleRequest(ctx context.Context, event json.RawMessage) (events.LambdaFunctionURLResponse, error) {
	// Unmarshal the outer event into a generic map.
	var genericEvent map[string]interface{}
	if err := json.Unmarshal(event, &genericEvent); err != nil {
		return buildResponse(400, fmt.Sprintf("Failed to parse event: %v", err)), nil
	}

	// Check for OPTIONS preflight by inspecting requestContext.http.method.
	if rc, exists := genericEvent["requestContext"].(map[string]interface{}); exists {
		if httpData, ok := rc["http"].(map[string]interface{}); ok {
			if method, ok := httpData["method"].(string); ok && method == "OPTIONS" {
				return buildResponse(200, "OK"), nil
			}
		}
	}

	// Extract and parse the inner payload from the "body" field.
	var innerPayload map[string]interface{}
	if body, exists := genericEvent["body"].(string); exists && body != "" {
		if err := json.Unmarshal([]byte(body), &innerPayload); err != nil {
			return buildResponse(400, fmt.Sprintf("Failed to parse body: %v", err)), nil
		}
	} else {
		return buildResponse(400, "Missing body in event"), nil
	}

	// Check if this is a content generation event (e.g. from EventBridge)
	if source, exists := genericEvent["source"]; exists && source == "aws.events" {
		fmt.Println("Event from EventBridge: processing content generation")
		if err := handleContentGeneration(ctx); err != nil {
			return buildResponse(500, fmt.Sprintf("Content generation error: %v", err)), nil
		}
		return buildResponse(200, "Content generation executed successfully."), nil
	}

	// Check if the inner payload contains an "action" field.
	if action, exists := innerPayload["action"].(string); exists {
		// Ensure the "email" field is present.
		email, emailExists := innerPayload["email"].(string)
		if !emailExists || email == "" {
			return buildResponse(400, "Email is required for subscription/unsubscription."), nil
		}

		if action == "subscribe" {
			fmt.Printf("Processing subscription for %s\n", email)
			if err := handleSubscription(ctx, email); err != nil {
				return buildResponse(500, fmt.Sprintf("Subscription error: %v", err)), nil
			}
			return buildResponse(200, "Subscription successful! Please check your email for confirmation."), nil
		} else if action == "unsubscribe" {
			fmt.Printf("Processing unsubscription for %s\n", email)
			if err := handleUnsubscription(ctx, email); err != nil {
				return buildResponse(500, fmt.Sprintf("Unsubscription error: %v", err)), nil
			}
			return buildResponse(200, "Unsubscription successful!"), nil
		} else {
			return buildResponse(400, fmt.Sprintf("Unknown action: %s", action)), nil
		}
	}

	// Fail if neither "action" nor "source" is provided.
	return buildResponse(400, "Missing required fields: either 'action' or 'source' must be provided."), nil
}

// handleSubscription processes a subscription request.
func handleSubscription(ctx context.Context, email string) error {
	fmt.Printf("Handling subscription for %s\n", email)
	err := helpers.SubscribeUser(ctx, helpers.SnsTopicARNHardcoded, email)
	if err != nil {
		return fmt.Errorf("subscription error: %w", err)
	}
	fmt.Println("Subscription successful!")
	return nil
}

// handleUnsubscription processes an unsubscription request.
func handleUnsubscription(ctx context.Context, email string) error {
	fmt.Printf("Handling unsubscription for %s\n", email)
	msg, err := helpers.UnsubscribeUser(ctx, helpers.SnsTopicARNHardcoded, email)
	if err != nil {
		return fmt.Errorf("unsubscription error: %w", err)
	}
	fmt.Println(msg)
	return nil
}

// handleContentGeneration processes the content generation workflow.
func handleContentGeneration(ctx context.Context) error {
	fmt.Println("Handling content generation event")

	// Retrieve secrets (NewsAPI & OpenAI keys)
	newsAPIKey, openaiAPIKey, err := helpers.GetSecrets(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving secrets: %w", err)
	}

	// Retrieve recent articles from DynamoDB.
	recentMap, err := helpers.GetRecentArticleURLs(ctx)
	if err != nil {
		fmt.Println("Error fetching recent articles from DynamoDB:", err)
	}

	// Accumulate valid articles.
	validArticles, err := helpers.AccumulateValidArticles(ctx, newsAPIKey, recentMap)
	if err != nil {
		return fmt.Errorf("error accumulating valid articles: %w", err)
	}
	fmt.Printf("Total valid articles accumulated: %d\n", len(validArticles))

	// Rank articles using GPT-4.
	openaiClient := openai.NewClient(openaiAPIKey)
	rankedArticles, err := helpers.RankArticlesWithChatGPT(ctx, openaiClient, validArticles)
	if err != nil {
		return fmt.Errorf("error ranking articles: %w", err)
	}
	fmt.Println("Ranking from GPT-4:")
	for _, ra := range rankedArticles {
		fmt.Printf("Rank %d: %s (%s) - Category: %s\n", ra.Rank, ra.Title, ra.URL, ra.Category)
	}

	// Select top articles (up to 10).
	topArticles := helpers.SelectTopArticles(rankedArticles, validArticles)
	if len(topArticles) > 10 {
		topArticles = topArticles[:10]
	}

	// Generate a pre-signed URL for latest_news.json.
	preSignedURL, err := helpers.GeneratePreSignedURL(ctx)
	if err != nil {
		fmt.Println("Error generating pre-signed URL:", err)
		preSignedURL = "Unavailable"
	}

	// Update index.html with the new pre-signed URL.
	if err := helpers.UpdateIndexHTML(ctx, preSignedURL); err != nil {
		fmt.Println("Error updating index.html:", err)
	}

	// Build the email message using BuildPlainMessage.
	plainMessage := helpers.BuildPlainMessage(topArticles, preSignedURL)

	// Send the email via SNS.
	if err := helpers.SendEmail(ctx, helpers.SnsTopicARNHardcoded, "Your Daily Uplifting News", plainMessage); err != nil {
		return fmt.Errorf("error sending email via SNS: %w", err)
	}

	fmt.Println("Content generation and email delivery completed successfully!")
	return nil
}

// buildResponse creates a LambdaFunctionURLResponse with CORS headers.
func buildResponse(status int, body string) events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
		},
		Body: fmt.Sprintf(`{"message": "%s"}`, body),
	}
}

func main() {
	lambda.Start(handleRequest)
}
