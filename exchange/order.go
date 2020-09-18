package exchange

import "github.com/shopspring/decimal"

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
	Id            uint64 // Id should be uint64
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
	Id          uint64 // Id should be uint64
	Price       decimal.Decimal
	Amount      decimal.Decimal
	FeeCurrency string
	FeeAmount   decimal.Decimal
}
