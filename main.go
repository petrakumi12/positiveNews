// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"positive-news/helpers"

	"github.com/aws/aws-lambda-go/lambda"
	openai "github.com/sashabaranov/go-openai"
)

func handleRequest(ctx context.Context, event json.RawMessage) error {
	// Retrieve API keys from Secrets Manager.
	newsAPIKey, openaiAPIKey, err := helpers.GetSecrets(ctx)
	if err != nil {
		fmt.Println("Error retrieving secrets:", err)
		return err
	}
	snsTopicARN := helpers.SnsTopicARNHardcoded

	// Retrieve recent article URLs from DynamoDB.
	recentMap, err := helpers.GetRecentArticleURLs(ctx)
	if err != nil {
		fmt.Println("Error fetching recent articles from DynamoDB:", err)
	}

	// Accumulate valid articles.
	validArticles, err := helpers.AccumulateValidArticles(ctx, newsAPIKey, recentMap)
	if err != nil {
		fmt.Println("Error accumulating valid articles:", err)
		return err
	}
	fmt.Printf("Total valid articles accumulated: %d\n", len(validArticles))

	// Rank articles using GPT-4.
	openaiClient := openai.NewClient(openaiAPIKey)
	rankedArticles, err := helpers.RankArticlesWithChatGPT(ctx, openaiClient, validArticles)
	if err != nil {
		fmt.Println("Error ranking articles:", err)
		return err
	}
	fmt.Println("Ranking from GPT-4:")
	for _, ra := range rankedArticles {
		fmt.Printf("Rank %d: %s (%s) - Category: %s\n", ra.Rank, ra.Title, ra.URL, ra.Category)
	}

	// Select top articles.
	topArticles := helpers.SelectTopArticles(rankedArticles, validArticles)

	// Store top articles in DynamoDB.
	if err := helpers.StoreArticles(ctx, topArticles); err != nil {
		fmt.Println("Error storing articles in DynamoDB:", err)
	} else {
		fmt.Println("Top articles stored successfully!")
	}

	// Build and send the SNS email.
	plainSubject := "Your Daily Positive News Rankings"
	plainMessage := helpers.BuildPlainMessage(topArticles, rankedArticles)
	if err := helpers.SendEmailViaSNS(ctx, snsTopicARN, plainSubject, plainMessage); err != nil {
		fmt.Println("Error sending email via SNS:", err)
		return err
	}
	fmt.Println("Email sent successfully via SNS!")
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
