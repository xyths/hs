package exchange

import "github.com/shopspring/decimal"

type Ticker struct {
	Timestamp int64 // unix timestamp in seconds
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
}

type Candle struct {
	Tickers []Ticker
}

const (
	GateIO = "gate"
	MXC    = "mxc"
	OKEx   = "okex"
	Huobi  = "huobi"
)