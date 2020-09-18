package huobi

const (
	OrderTypeBuyMarket  = "buy-market"
	OrderTypeSellMarket = "sell-market"
	OrderTypeBuyLimit   = "buy-limit"
	OrderTypeSellLimit  = "sell-limit"

	// IOC 全名 Immediate or Cancel（立即成交或撤单），
	// 指用户下的限价单，如果在指定价格不能立即成交，则撤销未成交的部分。
	OrderTypeBuyIoc  = "buy-ioc"
	OrderTypeSellIoc = "sell-ioc"

	// buy-limit-maker
	// 当“下单价格”>=“市场最低卖出价”，订单提交后，系统将拒绝接受此订单；
	// 当“下单价格”<“市场最低卖出价”，提交成功后，此订单将被系统接受。
	OrderTypeBuyLimitMaker = "buy-limit-maker"

	// sell-limit-maker
	// 当“下单价格”<=“市场最高买入价”，订单提交后，系统将拒绝接受此订单；
	// 当“下单价格”>“市场最高买入价”，提交成功后，此订单将被系统接受。
	OrderTypeSellLimitMaker = "sell-limit-maker"

	// 用户在下止盈止损订单时，须额外填写触发价 “stop-price” 与触发价运算符“operator”，
	// 并在订单类型 “type” 中指定订单类型 – “buy-stop-limit” 或 “sell-stop-limit”。
	// 其中，触发价运算符为 ”gte” 时，表示当市场最新成交价大于等于此触发价时该止盈止损订单将被触发；
	// 触发价运算符为 ”lte” 时，表示当市场最新成交价小于等于此触发价时该止盈止损订单将被触发。
	// 如果用户设置的触发价及运算符导致下单即被触发，该止盈止损订单将被拒绝接受。
	OrderTypeBuyStopLimit  = "buy-stop-limit"
	OrderTypeSellStopLimit = "sell-stop-limit"

	// - buy-limit-fok（FOK限价买单）
	// - sell-limit-fok（FOK限价卖单）
	// - buy-stop-limit-fok（FOK止盈止损限价买单）
	// - sell-stop-limit-fok（FOK止盈止损限价卖单）
	// 四个订单类型的订单有效期均为FOK（Fill or Kill），
	// 意即 – 如该订单下单后不能立即完全成交，则将被立即全部自动撤销。
	OrderTypeBuyLimitFok      = "buy-limit-fok"
	OrderTypeSellLimitFok     = "sell-limit-fok"
	OrderTypeBuyStopLimitFok  = "buy-stop-limit-fok"
	OrderTypeSellStopLimitFok = "sell-stop-limit-fok"
)
