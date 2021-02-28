package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/satori/go.uuid"
	"os"
	"strconv"
	"time"
)

var (
	desiredAsset = "USDT"
	btcSymbol = "BTC"
)

var prices []*binance.SymbolPrice
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
	CreatedAt    int64  `json:"created_at"`
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
	client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
	}

	var (
		tableName = aws.String(os.Getenv("BALANCE_TABLE_NAME"))
	)

	prices, _ = client.NewListPricesService().Do(context.Background())

	wallet := getWallet(account)
	// Write to DynamoDB
	item, _ := dynamodbattribute.MarshalMap(wallet)
	input := &dynamodb.PutItemInput{
		Item: item,
		TableName: tableName,
	}

	if _, err := ddb.PutItem(input); err != nil {
		return events.APIGatewayProxyResponse{ // Error HTTP response
			Body: err.Error(),
			StatusCode: 500,
		}, nil
	}

	js, err := json.Marshal(wallet)
	return events.APIGatewayProxyResponse{Body: string(js), StatusCode: 200}, nil
}

func getWallet(account *binance.Account) wallet {
	var walletBalance float64
	var wallet wallet
	id := uuid.Must(uuid.NewV4(), nil).String()
	for i := 0; i < len(account.Balances); i++ {

		if account.Balances[i].Asset == desiredAsset {
			walletBalance = walletBalance + convertAndSumStringValuesAsFloat(account.Balances[i].Free, account.Balances[i].Locked)
			continue
		}

		convertToBtc(account.Balances[i])
	}
	btcUsdParityPrice, _ := getTickerPrice(btcSymbol + desiredAsset)
	walletBalance = walletBalance + (walletBalanceAsBtc * btcUsdParityPrice)
	wallet.ID = id
	wallet.BalanceAsUSD = walletBalance
	wallet.BalanceAsBTC = walletBalanceAsBtc
	wallet.Coins = coins
	wallet.BtcUsd = btcUsdParityPrice
	wallet.CreatedAt = time.Now().Unix()
	fmt.Printf("Wallet Balance As Btc : %f\n", walletBalanceAsBtc)

	return wallet
}

func convertToBtc(balance binance.Balance) {
	total := convertAndSumStringValuesAsFloat(balance.Free, balance.Locked)
	totalAsBtc := total

	if balance.Asset != btcSymbol {
		tickerPriceAsBtc, _ := getTickerPrice(balance.Asset + btcSymbol)
		totalAsBtc = total * tickerPriceAsBtc
	}

	if totalAsBtc > 0 {
		coins = append(coins, coin{Asset: balance.Asset, Balance: total, TotalAsBtc: totalAsBtc})
		walletBalanceAsBtc = walletBalanceAsBtc + totalAsBtc
	}
}

func convertAndSumStringValuesAsFloat(num1 string, num2 string) float64 {
	floatNum1, _ := strconv.ParseFloat(num1, 32)
	floatNum2, _ := strconv.ParseFloat(num2, 32)
	total := floatNum1 + floatNum2

	return total
}

func getTickerPrice(parity string) (idx float64, err error) {
	for _, v := range prices {
		if v.Symbol == parity {
			return strconv.ParseFloat(v.Price, 32)
		}
	}

	return 1, err
}

