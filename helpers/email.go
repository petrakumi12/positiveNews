// email.go
package helpers

import (
	"fmt"
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
