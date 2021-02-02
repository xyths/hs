package exchange

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"time"
)

// common exchange interface for all symbols
type RestAPIExchange interface {
	FormatSymbol(base, quote string) string
	AllSymbols(ctx context.Context) (s []Symbol, err error)
	GetSymbol(ctx context.Context, symbol string) (Symbol, error)
	GetFee(symbol string) (fee Fee, err error)
	SpotBalance() (map[string]decimal.Decimal, error)
	SpotAvailableBalance() (map[string]decimal.Decimal, error)
	LastPrice(symbol string) (decimal.Decimal, error)
	Last24hVolume(symbol string) (decimal.Decimal, error)
	CandleBySize(symbol string, period time.Duration, size int) (hs.Candle, error)
	CandleFrom(symbol, clientId string, period time.Duration, from, to time.Time) (hs.Candle, error)

	//PlaceOrder(orderType, symbol, clientOrderId string, price, amount decimal.Decimal) (uint64, error)
	BuyLimit(symbol, clientOrderId string, price, amount decimal.Decimal) (orderId uint64, err error)
	SellLimit(symbol, clientOrderId string, price, amount decimal.Decimal) (orderId uint64, err error)
	BuyMarket(symbol Symbol, clientOrderId string, total decimal.Decimal) (orderId uint64, err error)
	SellMarket(symbol Symbol, clientOrderId string, amount decimal.Decimal) (orderId uint64, err error)
	BuyStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error)
	SellStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error)

	GetOrderById(orderId uint64, symbol string) (Order, error)
	CancelOrder(symbol string, orderId uint64) error
	IsFullFilled(symbol string, orderId uint64) (Order, bool, error)
}

type ResponseHandler func(response interface{})

type WsAPIExchange interface {
	SubscribeOrder(symbol, clientId string, responseHandler ResponseHandler)
	UnsubscribeOrder(symbol, clientId string)
	SubscribeCandlestick(symbol, clientId string, period time.Duration, responseHandler ResponseHandler)
	UnsubscribeCandlestick(symbol, clientId string, period time.Duration)
	SubscribeCandlestickWithReq(symbol, clientId string, period time.Duration, responseHandler ResponseHandler)
	UnsubscribeCandlestickWithReq(symbol, clientId string, period time.Duration)
}

// common exchange interface, for all symbols, all crypto-exchanges
type Exchange interface {
	RestAPIExchange
	WsAPIExchange
}

type Balance struct {
	Currency  string
	Available decimal.Decimal
	Locked    decimal.Decimal
}

type AccountBalance struct {
	Type       string // spot, margin, future
	Currencies []Balance
}
