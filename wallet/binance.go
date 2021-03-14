package main

import (
	"context"
	"fmt"
	binance "github.com/Akagi201/cryptotrader/binance"
	"github.com/Akagi201/cryptotrader/model"
	uuid "github.com/satori/go.uuid"
	"os"
	"time"
)

var prices []model.SimpleTicker
var client *binance.Client

func init() {
	apiKey, apiKeyPresent := os.LookupEnv("BINANCE_API_KEY")
	secretKey, secretKeyPresent := os.LookupEnv("BINANCE_SECRET_KEY")

	if !apiKeyPresent || !secretKeyPresent {
		fmt.Print("Binance keys are not present!")
		return
	}

	client = binance.New(apiKey, secretKey)
	prices, _ = client.GetTickers(context.Background())
}

func retrieveBalances() []model.Balance {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	balances, err := client.GetAccount(ctx, 0)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to get account from Binance: %s", err.Error()))
	}
	return balances
}

func getWallet() wallet {
	var walletBalance float64
	var wallet wallet
	id := uuid.Must(uuid.NewV4(), nil).String()

	var balances []model.Balance = retrieveBalances()

	for i := 0; i < len(balances); i++ {
		_, found := Find(usdCoins, balances[i].Currency)
		if found {
			walletBalance = walletBalance + balances[i].Free + balances[i].Frozen
			continue
		}

		addToWallet(balances[i])
	}
	btcUsdParityPrice, _ := getTickerPrice(btcSymbol + usdCoins[0])
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

func addToWallet(balance model.Balance) {
	total := balance.Free + balance.Frozen
	totalAsBtc := total

	if totalAsBtc <= 0 {
		return
	}

	if balance.Currency != btcSymbol {
		tickerPriceAsBtc, _ := getTickerPrice(balance.Currency + btcSymbol)
		totalAsBtc = total * tickerPriceAsBtc
	}

	coins = append(coins, coin{Asset: balance.Currency, Balance: total, TotalAsBtc: totalAsBtc})
	walletBalanceAsBtc = walletBalanceAsBtc + totalAsBtc
}

func getTickerPrice(parity string) (idx float64, err error) {
	for _, v := range prices {
		if v.Symbol == parity {
			return v.Price, nil
		}
	}

	return 0, err
}

func Find(slice [2]string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
