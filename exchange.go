package hs

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

// common exchange interface for all symbols
type RestAPIExchange interface {
	PricePrecision(symbol string) int32
	AmountPrecision(symbol string) int32
	MinAmount(symbol string) decimal.Decimal
	MinTotal(symbol string) decimal.Decimal

	SpotAvailableBalance() (map[string]decimal.Decimal, error)
	LastPrice(symbol string) (decimal.Decimal, error)
	CandleBySize(symbol string, period time.Duration, size int) (Candle, error)
	CandleFrom(symbol, clientId string, period time.Duration, from, to time.Time) (Candle, error)

	//PlaceOrder(orderType, symbol, clientOrderId string, price, amount decimal.Decimal) (int64, error)
	BuyLimit(symbol, clientOrderId string, price, amount decimal.Decimal) (orderId int64, err error)
	SellLimit(symbol, clientOrderId string, price, amount decimal.Decimal) (orderId int64, err error)
	BuyMarket(symbol, clientOrderId string, amount decimal.Decimal) (orderId int64, err error)
	SellMarket(symbol, clientOrderId string, amount decimal.Decimal) (orderId int64, err error)
	BuyStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId int64, err error)
	SellStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId int64, err error)

	GetOrderById(orderId uint64, symbol string) (Order, error)
	CancelOrder(symbol string, orderId uint64) error
}

type ResponseHandler func(response interface{})

type WsAPIExchange interface {
	SubscribeOrder(ctx context.Context, symbol, clientId string, responseHandler ResponseHandler)
	SubscribeCandlestick(ctx context.Context, symbol, clientId string, period time.Duration, responseHandler ResponseHandler)
	SubscribeCandlestickWithReq(ctx context.Context, symbol, clientId string, period time.Duration, responseHandler ResponseHandler)
}

// common exchange interface, for all symbols, all crypto-exchanges
type Exchange interface {
	RestAPIExchange
	WsAPIExchange
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
	Id            uint64
	ClientOrderId string `bson:"clientOrderId"`

	Type          string
	Symbol        string
	InitialPrice  decimal.Decimal `bson:"initialPrice"`
	InitialAmount decimal.Decimal
	Timestamp     int64

	Status string

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
