package hs

import (
	"context"
	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
	"github.com/shopspring/decimal"
	"time"
)

// common exchange interface for all symbols
type RestAPI interface {
	PricePrecision(symbol string) int32
	AmountPrecision(symbol string) int32
	GetSpotBalance() (map[string]decimal.Decimal, error)
	GetPrice(symbol string) (decimal.Decimal, error)
	GetCandle(symbol, clientId, period string, from, to time.Time) (Candle, error)

	PlaceOrder(orderType, symbol, clientOrderId string, price, amount decimal.Decimal) (uint64, error)
	CancelOrder(orderId uint64) error
}

type WsAPI interface {
	SubscribeCandlestick(ctx context.Context, symbol, clientId string, period time.Duration, responseHandler websocketclientbase.ResponseHandler)
	SubscribeCandlestickWithReq(ctx context.Context, symbol, clientId string, period time.Duration, responseHandler websocketclientbase.ResponseHandler)
}

// common exchange interface, for all symbols, all crypto-exchanges
type Exchange interface {
	RestAPI
	WsAPI
}

type OrderType = int

const (
	Buy  OrderType = 1
	Sell           = -1
)

type OrderStatus = int

const (
	Open      OrderStatus = 1 // open but not filled
	Closed                = 2 // full filled
	Filled                = 3 // part filled
	Cancelled             = 4
)

type Order struct {
	Id       uint64
	ClientId string

	Type          OrderType
	Symbol        string
	InitialPrice  decimal.Decimal
	InitialAmount decimal.Decimal
	Timestamp     int64

	Status OrderStatus

	FilledPrice  decimal.Decimal
	FilledAmount decimal.Decimal
	Trades       []Trade
	Fee          map[string]decimal.Decimal
}

type Trade struct {
	Id          uint64
	Price       decimal.Decimal
	Amount      decimal.Decimal
	FeeCurrency string
	FeeAmount   decimal.Decimal
}

const (
	MIN1  = time.Minute
	MIN5  = time.Minute * 5
	MIN15 = time.Minute * 15
	MIN30 = time.Minute * 30
	HOUR1 = time.Hour
	HOUR4 = time.Hour * 4
	DAY1  = time.Hour * 24
	MON1  = DAY1 * 30
	WEEK1 = DAY1 * 7
	YEAR1 = DAY1 * 365
)
