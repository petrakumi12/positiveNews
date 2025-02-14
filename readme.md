

# Positive News Aggregator
A tool that sends a daily email of feel-good news articles. 

This is a serverless Go application that aggregates optimistic news articles, ranks them using GPT-4, stores top articles in DynamoDB, and sends a daily email via SNS with your top 10 positive news articles of the past week.

## Overview

This project:

- **Fetches News:** Retrieves news articles using the NewsAPI.
- **Filters Articles:** Filters out articles with fewer than 150 words and those that have been sent in the past month (using DynamoDB).
- **Ranks Articles:** Uses GPT-4 (via the OpenAI API) to rank articles by positivity.
- **Stores Articles:** Saves the top articles in a DynamoDB table.
- **Sends Email:** Sends a plain text email via SNS with the top 10 positive articles.
- **Runs on Daily Schedule:** Designed to run as a Lambda function, triggered by an EventBridge rule on a daily schedule.

## Tools Used
	- **AWS Lambda** – Runs the function on a daily schedule.
	- **Amazon EventBridge** – Triggers the Lambda function at 7:20 AM PDT.
	- **AWS Secrets Manager** – Stores API keys securely.
	- **AWS DynamoDB** – Stores previously sent articles to prevent duplicates.
	- **AWS SNS** – Sends the top 10 articles via email.
	- **NewsAPI** – Fetches news articles based on positive keywords.
	- **OpenAI API** – Ranks articles by positivity.
	- **Go (Golang)** – Used for efficient and concurrent execution.
	- **AWS SAM CLI & Docker** – Enables local testing of the Lambda function.

## Code Workflow
	1.	Retrieve Secrets – Fetch API keys from AWS Secrets Manager.
	2.	Fetch News – Get articles from NewsAPI, handling pagination to avoid duplicates.
	3.	Filter Articles – Remove articles with <150 words and those recently sent.
	4.	Extract Content – Download and summarize article body (first 50 words).
	5.	Rank with GPT-4 – Analyze and rank the top 30 articles.
	6.	Store in DynamoDB – Save selected articles to prevent resending.
	7.	Send Email via SNS – Deliver the top 10 articles to subscribers.
	8.	Schedule Execution – AWS EventBridge triggers this workflow daily.


## Local Testing Using AWS SAM CLI

1. **Install SAM CLI:**  
   Follow the [SAM CLI installation guide](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).

2. **Create a `template.yaml`:**  
   For example:

   ```yaml
   AWSTemplateFormatVersion: '2010-09-09'
   Transform: AWS::Serverless-2016-10-31
   Resources:
     PositiveNewsFunction:
       Type: AWS::Serverless::Function
       Properties:
         Handler: main
         Runtime: go1.x
         CodeUri: .
         MemorySize: 512
         Timeout: 300
         Environment:
           Variables:
             SECRETS_MANAGER_SECRET_NAME: "positiveNews_openai_newsapi_keys"
         Policies:
           - SecretsManagerReadWritePolicy:
               SecretId: "positiveNews_openai_newsapi_keys"
           - DynamoDBCrudPolicy:
               TableName: "PositiveArticles"
           - SNSPublishMessagePolicy:
               TopicName: "positive_news"

3. **Build and Test**
- Build Go Project after major changes
```
rm go.sum && go clean -cache -modcache -testcache -x  && go mod tidy && go build 
```
- Build container and run test using SAM
```
GOOS=linux GOARCH=amd64 go build -o main && sam build --cached --use-container && sam build && sam local invoke OptimisticNewsFunction --event event.json
```


## Deploy to Lambda
- Build the project
```
GOOS=linux GOARCH=amd64 go build -o main
```
- Zip it
```
zip sendPositiveNews.zip main
```
- Update through AWS CLI
```
aws lambda update-function-code --function-name sendPositiveNews --zip-file fileb:///Users/petrakumi/workplace/positive-news/sendPositiveNews.zip
```

## Future Improvements
- Deduplication Across Sources:
    Enhance the logic to detect duplicate articles from different news sources (e.g., via normalized titles or content hashes).
- Daily Website Update:
    Generate and host a daily-updated website (using S3/CloudFront or AWS Amplify) with the full list of articles, and include that link in the email.
- Enhanced Error Handling & Logging:
    Add more detailed error handling and logging for easier debugging and monitoring.
- Secret Rotation:
    Implement automatic secret rotation using AWS Secrets Manager’s built-in rotation features.
- Caching:
    Consider caching API responses to reduce the number of external API calls and improve performance.
- UI changes to improve website look 
- Subscribtion to different topics
- Support article photos in website