// ranking.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// rankArticlesWithChatGPT sends up to 30 articles to GPT-4 for ranking.
func rankArticlesWithChatGPT(ctx context.Context, client *openai.Client, articles []ArticleWithContent) ([]RankedArticle, error) {
	if len(articles) > 30 {
		articles = articles[:30]
	}
	prompt := "Below are up to 30 articles with their title, URL, and a short excerpt (first 50 words of the article body). " +
		"Please analyze them and rank the articles from most positive to least positive, ensuring that the reader feels optimistic about the world. " +
		"Important: Only include an article if it is clearly positive. If fewer than 10 articles are clearly positive, return only those; do not add negative articles just to fill a top 10 list.\n\n" +
		"Follow these instructions exactly:\n\n" +
		"1. Exclude any articles that are about shopping, commerce, or product sales.\n" +
		"2. If there are many articles focused on self growth, self improvement, or positive thinking (self-help topics), include no more than 3 of those.\n" +
		"3. For each article, assign a suitable category from the following: business, entertainment, general, health, science, sports, technology, finance, world, arts, lifestyle.\n" +
		"4. Ensure that the final output includes only articles that are clearly positive. If fewer than 10 articles are clearly positive, return only those.\n" +
		"5. Return only a JSON array (with as many elements as are clearly positive) without any additional text. " +
		"Each JSON object must have the following fields: `rank` (an integer from 1 to N), `title`, `url`, and `category`.\n\n" +
		"Return only the JSON without any additional text.\n\nArticles:\n"
	for i, art := range articles {
		prompt += fmt.Sprintf("%d. Title: %s\nURL: %s\nExcerpt: %s\n\n", i+1, art.Title, art.URL, art.Excerpt)
	}
	req := openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are an expert sentiment analyst who curates news to inspire global positivity.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	resultText := strings.TrimSpace(resp.Choices[0].Message.Content)
	resultText = strings.TrimPrefix(resultText, "```")
	resultText = strings.TrimSuffix(resultText, "```")
	resultText = strings.TrimSpace(resultText)
	start := strings.Index(resultText, "[")
	end := strings.LastIndex(resultText, "]")
	if start >= 0 && end >= 0 && end > start {
		resultText = resultText[start : end+1]
	}
	fmt.Println("Cleaned ChatGPT Output:")
	fmt.Println(resultText)
	var ranked []RankedArticle
	if err := json.Unmarshal([]byte(resultText), &ranked); err != nil {
		return nil, fmt.Errorf("failed to parse JSON ranking: %v\nRaw output: %s", err, resultText)
	}
	return ranked, nil
}
