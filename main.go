package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

var (
	desiredAsset = "USDT"
	btcSymbol = "BTC"
)

var prices []*binance.SymbolPrice
var coins []coin
var walletBalanceAsBtc float64

type coin struct {
	asset string
	balance float64
	totalAsBtc float64
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	client := binance.NewClient(os.Getenv("apiKey"), os.Getenv("secretKey"))
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}
	prices, _ = client.NewListPricesService().Do(context.Background())

	calcBalances(account)
}

func calcBalances(account *binance.Account) {
	var walletBalanceAsDesired float64

	for i := 0; i < len(account.Balances); i++ {

		if account.Balances[i].Asset == desiredAsset {
			walletBalanceAsDesired = walletBalanceAsDesired + convertAndSumStringValuesAsFloat(account.Balances[i].Free, account.Balances[i].Locked)
			continue
		}

		convertToBtc(account.Balances[i])
	}

	tickerPriceAsDesired, _ := getTickerPrice(btcSymbol + desiredAsset)
	walletBalanceAsDesired = walletBalanceAsDesired + (walletBalanceAsBtc * tickerPriceAsDesired)
	fmt.Printf("Wallet Balance As Btc : %f\n", walletBalanceAsBtc)
	fmt.Printf("Wallet Balance As USDT : %f\n", walletBalanceAsDesired)
}

func convertToBtc(balance binance.Balance) {
	total := convertAndSumStringValuesAsFloat(balance.Free, balance.Locked)
	totalAsBtc := total

	if balance.Asset != btcSymbol {
		tickerPriceAsBtc, _ := getTickerPrice(balance.Asset + btcSymbol)
		totalAsBtc = total * tickerPriceAsBtc
	}

	if totalAsBtc > 0 {
		coins = append(coins, coin{asset: balance.Asset, balance: total, totalAsBtc: totalAsBtc})
		walletBalanceAsBtc = walletBalanceAsBtc + totalAsBtc
		fmt.Printf("%s: %f\n", balance.Asset, totalAsBtc)
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

