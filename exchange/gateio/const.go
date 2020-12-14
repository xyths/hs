package gateio

const (
	DefaultMaker = 0.002
	DefaultTaker = 0.002
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

const (
	WsIntervalSecond  = 5
	WsReconnectSecond = 60
)
