package huobi

import "github.com/shopspring/decimal"

// Trade is the Data object in TradeClear websocket response.
type Trade struct {
	Symbol    string
	OrderId   int64
	OrderSide string
	OrderType string
	Aggressor bool

	Id     int64           // trade id
	Time   int64           // trade time
	Price  decimal.Decimal // trade price
	Volume decimal.Decimal // trade volume

	TransactFee       decimal.Decimal
	FeeDeduct         decimal.Decimal
	FeeDeductCurrency string
}
