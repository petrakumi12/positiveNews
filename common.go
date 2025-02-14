// common.go
package main

// Shared constants
const (
	newsAPIURL           = "https://newsapi.org/v2/everything"
	tableName            = "PositiveArticles"
	snsTopicARNHardcoded = "arn:aws:sns:us-east-2:969666470832:positive_news"
)

// Shared type definitions
type NewsResponse struct {
	Articles []Article `json:"articles"`
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type ArticleWithContent struct {
	Title   string
	URL     string
	Excerpt string
}

type RankedArticle struct {
	Rank     int    `json:"rank"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Category string `json:"category"`
}
