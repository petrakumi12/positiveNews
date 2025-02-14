// util.go
package main

import "fmt"

// selectTopArticles selects up to 10 top articles from the ranked articles.
func selectTopArticles(rankedArticles []RankedArticle, validArticles []ArticleWithContent) []ArticleWithContent {
	articleMap := make(map[string]ArticleWithContent)
	for _, art := range validArticles {
		articleMap[art.URL] = art
	}
	var topArticles []ArticleWithContent
	for _, ra := range rankedArticles {
		if art, ok := articleMap[ra.URL]; ok {
			topArticles = append(topArticles, art)
		}
		if len(topArticles) >= 10 {
			break
		}
	}
	return topArticles
}

// buildPlainMessage builds the plain text email message.
func buildPlainMessage(topArticles []ArticleWithContent, rankedArticles []RankedArticle) string {
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
