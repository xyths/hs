package exchange

import (
	"github.com/shopspring/decimal"
	"time"
)

const (
	MIN1  = time.Minute
	MIN5  = time.Minute * 5
	MIN15 = time.Minute * 15
	MIN30 = time.Minute * 30
	HOUR1 = time.Hour
	HOUR4 = time.Hour * 4
	DAY1  = time.Hour * 24
	MON1  = DAY1 * 30
	WEEK1 = DAY1 * 7
	YEAR1 = DAY1 * 365
)

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

// price, amount
type Quote [2]float64

type OrderBook struct {
	Id   int
	Asks []Quote // sell
	Bids []Quote // buy
}
