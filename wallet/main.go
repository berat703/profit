package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"os"
)

var (
	usdCoins  = [...]string{"USDT", "BUSD"}
	btcSymbol = "BTC"
)

var coins []coin
var walletBalanceAsBtc float64

type coin struct {
	Asset      string  `json:"asset"`
	Balance    float64 `json:"balance"`
	TotalAsBtc float64 `json:"total_as_btc"`
}

type wallet struct {
	ID           string  `json:"id"`
	Coins        []coin  `json:"coins"`
	BalanceAsUSD float64 `json:"balance_as_usd"`
	BalanceAsBTC float64 `json:"balance_as_btc"`
	BtcUsd       float64 `json:"btc_usd"`
	CreatedAt    int64   `json:"created_at"`
}

var ddb *dynamodb.DynamoDB

func init() {
	region := os.Getenv("AWS_REGION")
	if session, err := session.NewSession(&aws.Config{ // Use aws sdk to connect to dynamoDB
		Region: &region,
	}); err != nil {
		fmt.Println(fmt.Sprintf("Failed to connect to AWS: %s", err.Error()))
	} else {
		ddb = dynamodb.New(session) // Create DynamoDB client
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var (
		tableName = aws.String(os.Getenv("BALANCE_TABLE_NAME"))
	)

	wallet := getWallet()
	// Write to DynamoDB
	item, _ := dynamodbattribute.MarshalMap(wallet)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: tableName,
	}

	if _, err := ddb.PutItem(input); err != nil {
		return events.APIGatewayProxyResponse{ // Error HTTP response
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	js, _ := json.Marshal(wallet)
	return events.APIGatewayProxyResponse{Body: string(js), StatusCode: 200}, nil
}
