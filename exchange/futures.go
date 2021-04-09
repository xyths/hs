package exchange

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type Contract struct {
	// Futures contract
	// 合约标识
	Name string `json:"name,omitempty"`
	// Futures contract type
	// 合约类型, inverse - 反向合约, direct - 正向合约
	Type string `json:"type,omitempty"`
	// Multiplier used in converting from invoicing to settlement currency in quote futures
	// 双币种合约中，计价货币兑换为结算货币的乘数
	QuoteMultiplier string `json:"quoteMultiplier,omitempty"`
	// Minimum leverage
	LeverageMin string `json:"leverage_min,omitempty"`
	// Maximum leverage
	LeverageMax string `json:"leverage_max,omitempty"`
	// Maintenance rate of margin
	MaintenanceRate string `json:"maintenance_rate,omitempty"`
	// Mark price type, internal - based on internal trading, index - based on external index price
	MarkType string `json:"mark_type,omitempty"`
	// Current mark price
	MarkPrice string `json:"mark_price,omitempty"`
	// Current index price
	IndexPrice string `json:"index_price,omitempty"`
	// Last trading price
	LastPrice string `json:"last_price,omitempty"`
	// Maker fee rate, where negative means rebate
	MakerFeeRate string `json:"maker_fee_rate,omitempty"`
	// Taker fee rate
	TakerFeeRate string `json:"taker_fee_rate,omitempty"`
	// Minimum order price increment
	OrderPriceRound string `json:"order_price_round,omitempty"`
	// Minimum mark price increment
	MarkPriceRound string `json:"mark_price_round,omitempty"`
	// Current funding rate
	FundingRate string `json:"funding_rate,omitempty"`
	// Funding application interval, unit in seconds
	FundingInterval int32 `json:"funding_interval,omitempty"`
	// Next funding time
	FundingNextApply float64 `json:"funding_next_apply,omitempty"`
	// Risk limit base
	RiskLimitBase string `json:"risk_limit_base,omitempty"`
	// Step of adjusting risk limit
	RiskLimitStep string `json:"risk_limit_step,omitempty"`
	// Maximum risk limit the contract allowed
	RiskLimitMax string `json:"risk_limit_max,omitempty"`
	// Minimum order size the contract allowed
	OrderSizeMin int64 `json:"order_size_min,omitempty"`
	// Maximum order size the contract allowed
	OrderSizeMax int64 `json:"order_size_max,omitempty"`
	// deviation between order price and current index price. If price of an order is denoted as order_price, it must meet the following condition:      abs(order_price - mark_price) <= mark_price * order_price_deviate
	OrderPriceDeviate string `json:"order_price_deviate,omitempty"`
	// Referral fee rate discount
	RefDiscountRate string `json:"ref_discount_rate,omitempty"`
	// Referrer commission rate
	RefRebateRate string `json:"ref_rebate_rate,omitempty"`
	// Current orderbook ID
	OrderbookId int64 `json:"orderbook_id,omitempty"`
	// Current trade ID
	TradeId int64 `json:"trade_id,omitempty"`
	// Historical accumulation trade size
	TradeSize int64 `json:"trade_size,omitempty"`
	// Current total long position size
	PositionSize int64 `json:"position_size,omitempty"`
	// Configuration's last changed time
	ConfigChangeTime float64 `json:"config_change_time,omitempty"`
	// Contract is delisting
	// 合约下线中
	InDelisting bool `json:"in_delisting,omitempty"`
	// Maximum number of open orders
	// 最多挂单数量
	OrdersLimit int32 `json:"orders_limit,omitempty"`
}

type FuturesQuote struct {
	Price  float64
	Amount int64
}

type FuturesOrderbook struct {
	Asks []FuturesQuote
	Bids []FuturesQuote
}

func (o FuturesOrderbook) String() string {
	var ret string
	for _, ask := range o.Asks {
		ret += fmt.Sprintf("%f\t%d\n", ask.Price, ask.Amount)
	}
	ret += "------\n"
	for _, bid := range o.Bids {
		ret += fmt.Sprintf("%f\t%d\n", bid.Price, bid.Amount)
	}
	return ret
}

type FuturesTrade struct {
	// Trade ID
	Id uint64 `json:"id,omitempty"`
	// 个人成交记录有 OrderId
	OrderId uint64 `json:"orderId,omitempty"`
	// Trading time
	CreateTime time.Time `json:"createTime,omitempty"`
	// Futures contract
	Contract string `json:"contract,omitempty"`
	// Trading size
	Size int64 `json:"size,omitempty"`
	// Trading price
	Price decimal.Decimal `json:"price,omitempty"`
	// 个人成交记录有 Role, maker/taker
	Role string `json:"role,omitempty"`
}

type FuturesLiquidation struct {
	// Liquidation time
	Time time.Time `json:"time,omitempty"`
	// Futures contract
	Contract string `json:"contract,omitempty"`
	// Position leverage. Not returned in public endpoints.
	Leverage int `json:"leverage,omitempty"`
	// Position size
	Size int64 `json:"size,omitempty"`
	// Position margin. Not returned in public endpoints.
	Margin decimal.Decimal `json:"margin,omitempty"`
	// Average entry price. Not returned in public endpoints.
	EntryPrice decimal.Decimal `json:"entryPrice,omitempty"`
	// Liquidation price. Not returned in public endpoints.
	LiqPrice decimal.Decimal `json:"liqPrice,omitempty"`
	// Mark price. Not returned in public endpoints.
	MarkPrice decimal.Decimal `json:"markPrice,omitempty"`
	// Liquidation order ID. Not returned in public endpoints.
	OrderId uint64 `json:"orderId,omitempty"`
	// Liquidation order price
	OrderPrice decimal.Decimal `json:"orderPrice,omitempty"`
	// Liquidation order average taker price
	FillPrice decimal.Decimal `json:"fillPrice,omitempty"`
	// Liquidation order maker size
	Left int64 `json:"left,omitempty"`
}

type FuturesBalance struct {
	// Total assets, total = position_margin + order_margin + available
	Total decimal.Decimal `json:"total,omitempty"`
	// Unrealized PNL
	UnrealisedPnl decimal.Decimal `json:"unrealisedPNL,omitempty"`
	// Position margin
	PositionMargin decimal.Decimal `json:"positionMargin,omitempty"`
	// Order margin of unfinished orders
	OrderMargin decimal.Decimal `json:"orderMargin,omitempty"`
	// Available balance to transfer out or trade
	Available decimal.Decimal `json:"available,omitempty"`
	// Settle currency
	Currency string `json:"currency,omitempty"`
	// Whether dual mode is enabled
	Dual bool `json:"dual,omitempty"`
}

// Current close order if any, or `null`
type PositionCloseOrder struct {
	// Close order ID
	// 委托ID
	Id int64 `json:"id,omitempty"`
	// Close order price
	// 委托价格
	Price decimal.Decimal `json:"price,omitempty"`
	// Is the close order from liquidation
	// 是否为强制平仓
	IsLiq bool `json:"isLiq,omitempty"`
}

// Futures position details
type Position struct {
	// User ID
	User int64 `json:"user,omitempty"`
	// Futures contract 合约标识
	Contract string `json:"contract,omitempty"`
	// Position size 头寸大小
	Size int64 `json:"size,omitempty"`
	// Position leverage. 0 means cross margin; positive number means isolated margin
	// 杠杆倍数，0代表全仓，正数代表逐仓
	Leverage int `json:"leverage,omitempty"`
	// Position risk limit 风险限额
	// 为了控制风险设置的持仓上限，若仓位上限增加，维持保证金和起始保证金要求也会提高
	RiskLimit int `json:"riskLimit,omitempty"`
	// Maximum leverage under current risk limit
	// 当前风险限额下，允许的最大杠杆倍数
	MaxLeverage int `json:"maxLeverage,omitempty"`
	// Maintenance rate under current risk limit
	// 当前风险限额下，维持保证金比例
	MaintenanceRate float64 `json:"maintenance_rate,omitempty"`
	// Position value calculated in settlement currency
	// 按结算币种标记价格计算的合约价值
	Value decimal.Decimal `json:"value,omitempty"`
	// Position margin 保证金
	Margin decimal.Decimal `json:"margin,omitempty"`
	// Entry price 开仓价格
	EntryPrice decimal.Decimal `json:"entry_price,omitempty"`
	// Liquidation price 爆仓价格
	LiqPrice decimal.Decimal `json:"liq_price,omitempty"`
	// Current mark price 合约当前标记价格
	MarkPrice decimal.Decimal `json:"mark_price,omitempty"`
	// Unrealized PNL 未实现盈亏
	UnrealisedPnl decimal.Decimal `json:"unrealised_pnl,omitempty"`
	// Realized PNL 已实现盈亏
	RealisedPnl decimal.Decimal `json:"realised_pnl,omitempty"`
	// History realized PNL 已平仓的仓位总盈亏
	HistoryPnl decimal.Decimal `json:"history_pnl,omitempty"`
	// PNL of last position close 最近一次平仓的盈亏
	LastClosePnl decimal.Decimal `json:"last_close_pnl,omitempty"`
	// ADL ranking, range from 1 to 5
	// 自动减仓排名，共1-5个等级
	AdlRanking int32 `json:"adl_ranking,omitempty"`
	// Current open orders 当前未完成委托数量
	PendingOrders int32 `json:"pending_orders,omitempty"`
	// 当前平仓委托信息，如果没有平仓则为null
	CloseOrder *PositionCloseOrder `json:"close_order,omitempty"`
	// Position mode, including:  - `single`: dual mode is not enabled- `dual_long`: long position in dual mode- `dual_short`: short position in dual mode
	// 持仓模式。包括：
	//   - single: 单向持仓模式
	//   - dual_long: 双向持仓模式下的做多仓位
	//   - dual_short: 双向持仓模式下的做空仓位
	Mode string `json:"mode,omitempty"`
}

// FuturesOrder, 期货单，即作为下单返回结果，也当作下单的入参。
type FuturesOrder struct {
	// Futures order ID
	Id uint64 `json:"id,omitempty"`
	// User ID
	User uint64 `json:"user,omitempty"`
	// Order creation time
	CreateTime time.Time `json:"create_time,omitempty"`
	// Order finished time. Not returned if order is open
	FinishTime time.Time `json:"finish_time,omitempty"`
	// How the order is finished.  - filled: all filled - cancelled: manually cancelled - liquidated: cancelled because of liquidation - ioc: time in force is `IOC`, finish immediately - auto_deleveraged: finished by ADL - reduce_only: cancelled because of increasing position while `reduce-only` set
	FinishAs string `json:"finish_as,omitempty"`
	// Order status  - `open`: waiting to be traded - `finished`: finished
	Status string `json:"status,omitempty"`
	// Futures contract
	Contract string `json:"contract"`
	// Order size. Specify positive number to make a bid, and negative number to ask
	Size int64 `json:"size"`
	// Display size for iceberg order. 0 for non-iceberg. Note that you would pay the taker fee for the hidden size
	Iceberg int64 `json:"iceberg,omitempty"`
	// Order price. 0 for market order with `tif` set as `ioc`
	Price decimal.Decimal `json:"price,omitempty"`
	// Set as `true` to close the position, with `size` set to 0
	Close bool `json:"close,omitempty"`
	// Is the order to close position
	IsClose bool `json:"is_close,omitempty"`
	// Set as `true` to be reduce-only order
	ReduceOnly bool `json:"reduce_only,omitempty"`
	// Is the order reduce-only
	IsReduceOnly bool `json:"is_reduce_only,omitempty"`
	// Is the order for liquidation
	IsLiq bool `json:"is_liq,omitempty"`
	// Time in force  - gtc: GoodTillCancelled - ioc: ImmediateOrCancelled, taker only - poc: PendingOrCancelled, reduce-only
	Tif string `json:"tif,omitempty"`
	// Size left to be traded
	Left int64 `json:"left,omitempty"`
	// Fill price of the order
	FillPrice decimal.Decimal `json:"fill_price,omitempty"`
	// User defined information. If not empty, must follow the rules below:  1. prefixed with `t-` 2. no longer than 28 bytes without `t-` prefix 3. can only include 0-9, A-Z, a-z, underscore(_), hyphen(-) or dot(.) Besides user defined information, reserved contents are listed below, denoting how the order is created:  - web: from web - api: from API - app: from mobile phones - auto_deleveraging: from ADL - liquidation: from liquidation - insurance: from insurance
	Text string `json:"text,omitempty"`
	// Taker fee rate
	TakerFee float64 `json:"takerFee,omitempty"`
	// Maker fee rate
	MakerFee float64 `json:"makerFee,omitempty"`
	// Reference user ID
	Reference uint64 `json:"reference,omitempty"`
}
