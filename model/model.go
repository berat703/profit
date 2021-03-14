package model

type Coin struct {
	Asset      string  `json:"asset"`
	Balance    float64 `json:"balance"`
	TotalAsBtc float64 `json:"total_as_btc"`
}

type Wallet struct {
	ID           string  `json:"id"`
	Coins        []Coin  `json:"coins"`
	BalanceAsUSD float64 `json:"balance_as_usd"`
	BalanceAsBTC float64 `json:"balance_as_btc"`
	BtcUsd       float64 `json:"btc_usd"`
	CreatedAt    int64   `json:"created_at"`
}
