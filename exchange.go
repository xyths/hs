package hs

import (
	"context"
	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
	"github.com/shopspring/decimal"
)

type Exchange interface {
	GetSpotBalance() (map[string]decimal.Decimal, error)

	PlaceOrder(orderType, clientOrderId string, price, amount decimal.Decimal) (uint64, error)
	CancelOrder(orderId uint64) error

	SubscribeCandlestick(ctx context.Context, symbol, clientId string, responseHandler websocketclientbase.ResponseHandler)
}
