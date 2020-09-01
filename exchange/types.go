package exchange

import "github.com/shopspring/decimal"

type Ticker struct {
	Id            int64
	Last          decimal.Decimal // 最新成交价
	LowestAsk     decimal.Decimal // 卖1，卖方最低价
	HighestBid    decimal.Decimal // 买1，买方最高价
	PercentChange decimal.Decimal // 涨跌百分比
	BaseVolume    decimal.Decimal // 交易量
	QuoteVolume   decimal.Decimal // 兑换货币交易量
	High24hr      decimal.Decimal // 24小时最高价
	Low24hr       decimal.Decimal // 24小时最低价
}
