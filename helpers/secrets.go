// secrets.go
package helpers

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	sm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// getSecrets retrieves NEWS_API_KEY and OPENAI_API_KEY from AWS Secrets Manager.
func GetSecrets(ctx context.Context) (newsAPIKey, openaiAPIKey string, err error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", err
	}
	smClient := sm.NewFromConfig(cfg)
	secretName := "positiveNews_openai_newsapi_keys" // Hardcoded secret name.
	input := &sm.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := smClient.GetSecretValue(ctx, input)
	if err != nil {
		return "", "", err
	}
	var secretMap map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretMap); err != nil {
		return "", "", err
	}
	return secretMap["NEWS_API_KEY"], secretMap["OPENAI_API_KEY"], nil
}
