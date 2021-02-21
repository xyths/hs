package exchange

import (
	"github.com/shopspring/decimal"
	"time"
)

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

// Order is common order type between all exchanges, use for exchange interface
type Order struct {
	Id            uint64 `json:"id"` // Id should be uint64
	ClientOrderId string `json:"clientOrderId" bson:"clientOrderId"`

	// gate: limit
	// huobi:
	Type   string          `json:"type"`
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
	Amount decimal.Decimal `json:"amount"`
	Time   time.Time       `json:"time"`

	Status string `json:"status"`

	FilledPrice  decimal.Decimal `json:"filledPrice"`
	FilledAmount decimal.Decimal `json:"filledAmount"`
	Trades       []Trade         `json:"trades,omitempty"`
}

type Trade struct {
	Id      uint64 `json:"id"` // Id should be uint64
	OrderId uint64 `json:"orderId"`
	Symbol  string `json:"symbol,omitempty"`
	Type    string `json:"type,omitempty"`
	// v4, side = buy/sell, role = maker/taker
	Side        string          `json:"side,omitempty"`
	Role        string          `json:"role,omitempty"`
	Price       decimal.Decimal `json:"price"`
	Amount      decimal.Decimal `json:"amount"`
	FeeCurrency string          `json:"feeCurrency,omitempty"`
	FeeAmount   decimal.Decimal `json:"feeAmount,omitempty"`
	Time        time.Time       `json:"time"`
}
