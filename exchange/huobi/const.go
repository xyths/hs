package huobi

const (
	BTC_USDT = "btcusdt"
)

var (
	PricePrecision = map[string]int{
		"btcusdt": 2,
	}
	AmountPrecision = map[string]int{
		"btcusdt": 5,
	}
	MinAmount = map[string]float64{
		"btcusdt": 0.0001,
	}
	MinTotal = map[string]int64{
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
