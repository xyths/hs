package huobi

import "github.com/shopspring/decimal"

// Trade is the Data object in TradeClear websocket response.
type Trade struct {
	Symbol    string
	OrderId   int64
	OrderSide string
	OrderType string
	Aggressor bool

	Id          int64
	TradeTime   int64
	TradePrice  decimal.Decimal
	TradeVolume decimal.Decimal

	TransactFee       decimal.Decimal
	FeeDeduct         decimal.Decimal
	FeeDeductCurrency string
}
