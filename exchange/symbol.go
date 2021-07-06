package exchange

import "github.com/shopspring/decimal"

type Symbol struct {
	Symbol              string
	Disabled            bool
	BaseCurrency        string          `json:"baseCurrency"`  // 交易对中的基础币种, coin, eg. BTC
	QuoteCurrency       string          `json:"quoteCurrency"` // 交易对中的报价币种, cash, eg. USDT
	PricePrecision      int32           `json:"pricePrecision"`
	AmountPrecision     int32           `json:"amountPrecision"`
	LimitOrderMinAmount decimal.Decimal `json:"minAmount"`
	MinTotal            decimal.Decimal `json:"minTotal"`
}

type Fee struct {
	Symbol      string
	BaseMaker   decimal.Decimal `json:"baseMaker"`
	BaseTaker   decimal.Decimal `json:"baseTaker"`
	ActualMaker decimal.Decimal `json:"actualMaker"`
	ActualTaker decimal.Decimal `json:"actualTaker"`
}
