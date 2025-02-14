// news.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
)

// fetchNews retrieves up to 50 articles from NewsAPI for a given page.
func fetchNews(apiKey string, page int) ([]Article, error) {
	query := "inspiring OR heartwarming OR motivational OR encouraging OR breakthrough OR innovation OR success OR 'good news' OR uplifting OR inspiring -crisis -war -tragedy -disaster -shooting"
	today := time.Now()
	sevenDaysAgo := today.AddDate(0, 0, -7)
	fromDate := sevenDaysAgo.Format("2006-01-02")
	toDate := today.Format("2006-01-02")
	requestURL := fmt.Sprintf("%s?q=%s&from=%s&to=%s&sortBy=relevancy&pageSize=50&page=%d&language=en&apiKey=%s",
		newsAPIURL, url.QueryEscape(query), fromDate, toDate, page, apiKey)
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("NewsAPI Raw Output (page %d):\n", page)
	fmt.Println(string(body))
	var newsResp NewsResponse
	if err := json.Unmarshal(body, &newsResp); err != nil {
		return nil, err
	}
	return newsResp.Articles, nil
}

// fetchArticleContent downloads the article content from a URL.
func fetchArticleContent(articleURL string) (string, error) {
	resp, err := http.Get(articleURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return "", err
	}
	doc, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return "", err
	}
	return doc.TextContent, nil
}

// getFiftyWordExcerpt returns the first 50 words of the given text.
func getFiftyWordExcerpt(text string) string {
	words := strings.Fields(text)
	if len(words) < 50 {
		return ""
	}
	return strings.Join(words[:50], " ")
}

// accumulateValidArticles fetches and filters articles until up to 30 valid ones are accumulated.
func accumulateValidArticles(ctx context.Context, apiKey string, recentMap map[string]bool) ([]ArticleWithContent, error) {
	var validArticles []ArticleWithContent
	seen := make(map[string]bool)
	attempts := 0
	maxAttempts := 3
	page := 1
	for len(validArticles) < 30 && attempts < maxAttempts {
		articles, err := fetchNews(apiKey, page)
		if err != nil {
			return nil, err
		}
		for _, art := range articles {
			if seen[art.URL] {
				continue
			}
			if recentMap != nil {
				if _, exists := recentMap[art.URL]; exists {
					continue
				}
			}
			content, err := fetchArticleContent(art.URL)
			if err != nil {
				fmt.Printf("Error fetching content for article '%s': %v\n", art.Title, err)
				continue
			}
			words := strings.Fields(content)
			if len(words) < 150 {
				continue
			}
			excerpt := getFiftyWordExcerpt(content)
			if excerpt == "" {
				continue
			}
			validArticles = append(validArticles, ArticleWithContent{
				Title:   art.Title,
				URL:     art.URL,
				Excerpt: excerpt,
			})
			seen[art.URL] = true
			if len(validArticles) >= 30 {
				break
			}
		}
		attempts++
		page++
		if len(validArticles) < 30 && attempts < maxAttempts {
			fmt.Printf("Accumulated %d valid articles so far; fetching again (attempt %d of %d, page %d)...\n", len(validArticles), attempts+1, maxAttempts, page)
			time.Sleep(2 * time.Second)
		}
	}
	return validArticles, nil
}
