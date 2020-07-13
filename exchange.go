package hs

import (
	"context"
	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
	"github.com/shopspring/decimal"
)

type RestAPI interface {
	GetSpotBalance() (map[string]decimal.Decimal, error)

	PlaceOrder(orderType, symbol, clientOrderId string, price, amount decimal.Decimal) (uint64, error)
	CancelOrder(orderId uint64) error
}

type WsAPI interface {
	SubscribeCandlestick(ctx context.Context, symbol, clientId string, responseHandler websocketclientbase.ResponseHandler)
}

// common exchange interface, for all symbols, all crypto-exchanges
type Exchange interface {
	RestAPI
	WsAPI
}
