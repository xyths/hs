package exchange

import "github.com/shopspring/decimal"

// TradeDetail 是成交明细，不是个人账户成交历史，而是系统整体的成交明细（逐笔成交明细）。
// 这是统一各交易所差异后的结构。
type TradeDetail struct {
	Id        int64           `json:"id"` // trade id
	Price     decimal.Decimal `json:"price"`
	Amount    decimal.Decimal `json:"amount"`
	Timestamp int64           `json:"timestamp"`
	Direction Direction       `json:"direction"`
}

type Direction = string

// 火币的交易明细里直接使用了buy/sell，因此不需要转换。
const (
	TradeDirectionBuy  Direction = "buy"
	TradeDirectionSell           = "sell"
)

// TradeHandler 是订阅交易明细时的处理函数。
// 参数中的数据是按时间顺序排列，老数据在前，新数据在后，方便遍历和用TA分析。
type TradeHandler func([]TradeDetail)
