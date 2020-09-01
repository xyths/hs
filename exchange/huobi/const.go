package huobi

import "github.com/shopspring/decimal"

const (
	BTC_USDT = "btcusdt"
)

var (
	PricePrecision = map[string]int32{
		"btcusdt": 2,
	}
	AmountPrecision = map[string]int32{
		"btcusdt": 5,
	}
	MinAmount = map[string]float64{
		"btcusdt": 0.0001,
	}
	MinTotal = map[string]float64{
		"btcusdt": 5,
	}
)

const (
	OrderTypeBuyMarket        = "buy-market"
	OrderTypeSellMarket       = "sell-market"
	OrderTypeBuyLimit         = "buy-limit"
	OrderTypeSellLimit        = "sell-limit"
	OrderTypeBuyIoc           = "buy-ioc"
	OrderTypeSellIoc          = "sell-ioc"
	OrderTypeBuyLimitMaker    = "buy-limit-maker"
	OrderTypeSellLimitMaker   = "sell-limit-maker"
	OrderTypeBuyStopLimit     = "buy-stop-limit"
	OrderTypeSellStopLimit    = "sell-stop-limit"
	OrderTypeBuyLimitFok      = "buy-limit-fok"
	OrderTypeSellLimitFok     = "sell-limit-fok"
	OrderTypeBuyStopLimitFok  = "buy-stop-limit-fok"
	OrderTypeSellStopLimitFok = "sell-stop-limit-fok"
)

func (c Client) PricePrecision(symbol string) int32 {
	return PricePrecision[symbol]
}

func (c Client) AmountPrecision(symbol string) int32 {
	return AmountPrecision[symbol]
}

func (c Client) MinAmount(symbol string) decimal.Decimal {
	return decimal.NewFromFloat(MinAmount[symbol])
}

func (c Client) MinTotal(symbol string) decimal.Decimal {
	return decimal.NewFromFloat(MinTotal[symbol])
}
