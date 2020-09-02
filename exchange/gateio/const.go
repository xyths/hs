package gateio

import "github.com/shopspring/decimal"

const (
	BTC_USDT   = "btc_usdt"
	BTC3L_USDT = "btc3l_usdt"
	BTC3S_USDT = "btc3s_usdt"
	SERO_USDT  = "sero_usdt"

	BTC   = "BTC"
	BTC3L = "BTC3L"
	BTC3S = "BTC3S"
	USDT  = "USDT"
	SERO  = "SERO"
)

var (
	PricePrecision = map[string]int32{
		BTC_USDT:   2,
		BTC3L_USDT: 4,
		BTC3S_USDT: 4,
		SERO_USDT:  5,
	}
	AmountPrecision = map[string]int32{
		BTC_USDT:   4,
		BTC3L_USDT: 3,
		BTC3S_USDT: 3,
		SERO_USDT:  3,
	}
	MinAmount = map[string]float64{
		BTC_USDT:   0.0001,
		BTC3L_USDT: 0.001,
		BTC3S_USDT: 0.001,
		SERO_USDT:  0.001,
	}
	MinTotal = map[string]float64{
		BTC_USDT:   1,
		BTC3L_USDT: 1,
		BTC3S_USDT: 1,
		SERO_USDT:  1,
	}
)

// used by buy/sell
const (
	// 订单类型("gtc"：普通订单（默认）；
	// “ioc”：立即执行否则取消订单（Immediate-Or-Cancel，IOC）；
	// "poc":被动委托（只挂单，不吃单）（Pending-Or-Cancelled，POC）)
	OrderTypeNormal = "gtc"
	OrderTypeGTC    = "gtc"
	OrderTypeIOC    = "ioc"
	OrderTypePOC    = "poc"
)

const (
	OrderStatusOpen      = "open"
	OrderStatusCancelled = "cancelled"
	OrderStatusClosed    = "closed"

	OrderTypeBuy  = "buy"
	OrderTypeSell = "sell"
)

func (g GateIO) PricePrecision(symbol string) int32 {
	return PricePrecision[symbol]
}

func (g GateIO) AmountPrecision(symbol string) int32 {
	return AmountPrecision[symbol]
}

func (g GateIO) MinAmount(symbol string) decimal.Decimal {
	return decimal.NewFromFloat(MinAmount[symbol])
}

func (g GateIO) MinTotal(symbol string) decimal.Decimal {
	return decimal.NewFromFloat(MinTotal[symbol])
}