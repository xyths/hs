package gateio

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v5"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
	"github.com/xyths/hs/exchange"
	"go.uber.org/zap"
	"time"
)

// SpotV4 is gate.io API v4 wrapper
// GetXxx/Xxx: RESTful GET
// ReqXxx: websocket query
// SubXxx/UnsubXxx: websocket subscribe or unsubscribe
type SpotV4 struct {
	Key    string
	Secret string
	client *gateapi.APIClient
	wsHost string
	wsPath string

	Logger *zap.SugaredLogger
}

func NewSpotV4(key, secret, host string, logger *zap.SugaredLogger) *SpotV4 {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	return &SpotV4{Key: key, Secret: secret, client: client, wsHost: host, wsPath: "/v4", Logger: logger}
}

// class function layout
//  - public
//    - RESTful
//      - common
//      - spot
//    - ws
//      - common
//      - spot
//  - private
//    - RESTful
//      - common
//      - spot
//    - ws
//      - common
//      - spot

func (g *SpotV4) Name() string {
	return "gate"
}

// list all currencies
func (g *SpotV4) Currencies(ctx context.Context) ([]gateapi.Currency, error) {
	currencies, _, err := g.client.SpotApi.ListCurrencies(ctx)
	return currencies, err
}

// 获取一定数量的k线
// 可以用于策略启动时的查询
func (g *SpotV4) CandleBySize(symbol string, period time.Duration, size int) (hs.Candle, error) {
	return g.CandleBySizeContext(context.Background(), symbol, period, size)
}
func (g *SpotV4) CandleBySizeContext(ctx context.Context, symbol string, period time.Duration, size int) (hs.Candle, error) {
	interval := optional.NewString(getInterval(period))
	left := size
	to := time.Now()
	type param struct {
		Limit int
		To    time.Time
	}
	var params []param
	for left > 0 {
		limit := left
		if limit > maxCandleLength {
			limit = maxCandleLength
		}
		params = append(params, param{Limit: limit, To: to})
		left -= limit
		to = to.Add(period * time.Duration(-limit))
	}
	candle := hs.NewCandle(size)
	for i := len(params) - 1; i >= 0; i-- {
		options := &gateapi.ListCandlesticksOpts{
			Limit:    optional.NewInt32(int32(params[i].Limit)),
			To:       optional.NewInt64(params[i].To.Unix()),
			Interval: interval,
		}
		c, err := g.listCandlesticks(ctx, symbol, options)
		if err != nil {
			if candle.Length() > 0 {
				return candle, nil
			} else {
				return candle, err
			}
		}
		candle.Add(c)
		//if c.Length() < int(options.Limit.Value()) {
		//	break
		//}
	}
	return candle, nil
}

func (g *SpotV4) AllSymbols(ctx context.Context) (symbols []exchange.Symbol, err error) {
	pairs, _, err := g.client.SpotApi.ListCurrencyPairs(ctx)
	if err != nil {
		return
	}
	for _, p := range pairs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			s := exchange.Symbol{
				Symbol:              p.Id,
				Disabled:            p.TradeStatus != "tradable",
				BaseCurrency:        p.Base,
				QuoteCurrency:       p.Quote,
				PricePrecision:      p.Precision,
				AmountPrecision:     p.AmountPrecision,
				LimitOrderMinAmount: decimal.Zero,
				MinTotal:            decimal.Zero,
			}
			if ma, err1 := decimal.NewFromString(p.MinBaseAmount); err1 == nil {
				s.LimitOrderMinAmount = ma
			}
			if mt, err1 := decimal.NewFromString(p.MinQuoteAmount); err1 == nil {
				s.MinTotal = mt
			}
			symbols = append(symbols, s)
		}
	}
	return
}

func (g *SpotV4) GetSymbol(ctx context.Context, symbol string) (s exchange.Symbol, err error) {
	p, _, err := g.client.SpotApi.GetCurrencyPair(ctx, symbol)
	if err != nil {
		return
	}
	s.Symbol = p.Id
	s.Disabled = p.TradeStatus != "tradable"
	s.BaseCurrency = p.Base
	s.QuoteCurrency = p.Quote
	s.PricePrecision = p.Precision
	s.AmountPrecision = p.AmountPrecision
	if ma, err1 := decimal.NewFromString(p.MinBaseAmount); err1 == nil {
		s.LimitOrderMinAmount = ma
	} else {
		s.LimitOrderMinAmount = decimal.Zero
	}
	if mt, err1 := decimal.NewFromString(p.MinQuoteAmount); err1 == nil {
		s.MinTotal = mt
	} else {
		s.MinTotal = decimal.Zero
	}
	return
}

// Balance returns account balances
func (g *SpotV4) Balance(ctx context.Context) ([]exchange.Balance, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	spot, _, err := g.client.SpotApi.ListSpotAccounts(ctx2, nil)
	if err != nil {
		return nil, err
	}
	var result []exchange.Balance
	for _, currency := range spot {
		balance := exchange.Balance{
			Currency: currency.Currency,
		}
		if d, err1 := decimal.NewFromString(currency.Available); err1 == nil {
			balance.Available = d
		}
		if d, err1 := decimal.NewFromString(currency.Locked); err1 == nil {
			balance.Locked = d
		}
		result = append(result, balance)
	}
	return result, nil
}

// AvailableBalance returns account available balances
func (g *SpotV4) AvailableBalance(ctx context.Context) (map[string]decimal.Decimal, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	all, err := g.Balance(ctx2)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]decimal.Decimal)
	for _, c := range all {
		if c.Available.IsPositive() {
			balance[c.Currency] = c.Available
		}
	}
	return balance, nil
}

// 订单类型("gtc"：普通订单（默认）；“ioc”：立即执行否则取消订单（Immediate-Or-Cancel，IOC）；"poc":被动委托（只挂单，不吃单）（Pending-Or-Cancelled，POC）)
func (g *SpotV4) BuyLimit(ctx context.Context, symbol, clientOrderId string, price, amount decimal.Decimal) (exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	return g.placeOrder(ctx2, symbol, price, amount, "buy", OrderTypeGTC, clientOrderId)
}

// Place order sell
func (g *SpotV4) SellLimit(ctx context.Context, symbol, text string, price, amount decimal.Decimal) (exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	return g.placeOrder(ctx2, symbol, price, amount, "sell", OrderTypeGTC, text)
}

// BuyMarket use Ticker's last price to place order.
// may can't fill when big bull
func (g *SpotV4) BuyMarket(ctx context.Context, symbol exchange.Symbol, clientOrderId string, total decimal.Decimal) (exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	return g.buyFromOrderBook(ctx2, symbol, clientOrderId, total)
}

// SellMarket use Ticker's last price to place order.
// may can't fill when big bear
func (g *SpotV4) SellMarket(ctx context.Context, symbol exchange.Symbol, clientOrderId string, amount decimal.Decimal) (exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	return g.sellFromOrderBook(ctx2, symbol, clientOrderId, amount)
}

// list all orders by status (open, finished)
func (g *SpotV4) ListOrders(ctx context.Context, symbol, status string) ([]exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	orders, _, err := g.client.SpotApi.ListOrders(ctx2, symbol, status, &gateapi.ListOrdersOpts{})
	if err != nil {
		return nil, err
	}
	var ret []exchange.Order
	for _, o := range orders {
		ret = append(ret, convertOrder(o))
	}
	return ret, nil
}

// list open orders by symbol
func (g *SpotV4) ListOpenOrders(ctx context.Context, symbol string) ([]exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	return g.ListOrders(ctx2, symbol, "open")
}

func (g *SpotV4) GetOrder(ctx context.Context, symbol string, orderId uint64) (exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	raw, _, err := g.client.SpotApi.GetOrder(ctx2, symbol, fmt.Sprintf("%d", orderId))
	if err != nil {
		return exchange.Order{}, err
	}
	// status will be
	// open, closed, cancelled
	return convertOrder(raw), nil
}

func (g *SpotV4) IsFullFilled(ctx context.Context, symbol string, orderId uint64) (order exchange.Order, filled bool, err error) {
	order, err = g.GetOrder(ctx, symbol, orderId)
	if err != nil {
		return
	}
	filled = order.Status == "closed"
	return
}

// Cancel order
func (g *SpotV4) CancelOrder(ctx context.Context, symbol string, orderId uint64) (exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	raw, _, err := g.client.SpotApi.CancelOrder(ctx2, fmt.Sprintf("%d", orderId), symbol)
	if err != nil {
		return exchange.Order{}, err
	}
	return convertOrder(raw), nil
}

// cancel all orders
func (g *SpotV4) CancelAllOrders(ctx context.Context, symbol string) ([]exchange.Order, error) {
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	rawOrders, _, err := g.client.SpotApi.CancelOrders(ctx2, symbol, &gateapi.CancelOrdersOpts{Account: optional.NewString("spot")})
	if err != nil {
		return nil, err
	}
	var orders []exchange.Order
	for _, r := range rawOrders {
		orders = append(orders, convertOrder(r))
	}
	return orders, nil
}

// my trades history
func (g *SpotV4) MyTrades(ctx context.Context, symbol, orderId string) ([]exchange.Trade, error) {
	opts := gateapi.ListMyTradesOpts{}
	if orderId != "" {
		opts.OrderId = optional.NewString(orderId)
	}
	ctx2 := context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    g.Key,
		Secret: g.Secret,
	})
	rawTrades, _, err := g.client.SpotApi.ListMyTrades(ctx2, symbol, &opts)
	if err != nil {
		return nil, err
	}
	var trades []exchange.Trade
	for _, trade := range rawTrades {
		trades = append(trades, convertTrade(trade))
	}
	return trades, nil
}

// tickers
func (g *SpotV4) Tickers(ctx context.Context) ([]exchange.Ticker, error) {
	rawList, _, err := g.client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{})
	if err != nil {
		return nil, err
	}
	var tickers []exchange.Ticker
	for _, t := range rawList {
		tickers = append(tickers, convertTicker(t))
	}
	return tickers, nil
}

// ticker
func (g *SpotV4) Ticker(ctx context.Context, symbol string) (exchange.Ticker, error) {
	rawList, _, err := g.client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{CurrencyPair: optional.NewString(symbol)})
	if err != nil {
		return exchange.Ticker{}, err
	}
	if len(rawList) == 1 {
		return convertTicker(rawList[0]), nil
	} else {
		return exchange.Ticker{}, nil
	}
}

func (g *SpotV4) LastPrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
	ticker, err := g.Ticker(ctx, symbol)
	if err != nil {
		return decimal.Zero, err
	}
	return ticker.Last, nil
}

func (g *SpotV4) Last24hVolume(ctx context.Context, symbol string) (decimal.Decimal, error) {
	ticker, err := g.Ticker(ctx, symbol)
	if err != nil {
		return decimal.Zero, err
	}
	return ticker.BaseVolume, nil
}

// API限制最大数目是1000根
const maxCandleLength = 1000

func (g *SpotV4) listCandlesticks(ctx context.Context, symbol string, options *gateapi.ListCandlesticksOpts) (hs.Candle, error) {
	result, _, err := g.client.SpotApi.ListCandlesticks(ctx, symbol, options)
	if err != nil {
		return hs.Candle{}, err
	}
	candles := hs.NewCandle(len(result))
	for i := 0; i < len(result); i++ {
		c := result[i]
		candles.Append(hs.Ticker{
			Timestamp: convert.StrToInt64(c[0]),
			// volume count by quote currency, eg. BTC_USDT is xxx USDT
			Volume: convert.StrToFloat64(c[1]),
			Close:  convert.StrToFloat64(c[2]),
			High:   convert.StrToFloat64(c[3]),
			Low:    convert.StrToFloat64(c[4]),
			Open:   convert.StrToFloat64(c[5]),
		})
	}
	return candles, nil
}

// placeOrder is a internal function
// it convert struct gateio order to standard order type
func (g *SpotV4) placeOrder(ctx context.Context, symbol string, price, amount decimal.Decimal, side, orderType, text string) (exchange.Order, error) {
	orderRequest := gateapi.Order{
		Account:      "spot",
		Type:         "limit",
		CurrencyPair: symbol,
		Price:        price.String(),
		Amount:       amount.String(),
		Side:         side,
		TimeInForce:  orderType,
		Text:         fmt.Sprintf("t-%s", text),
	}
	r, _, err := g.client.SpotApi.CreateOrder(ctx, orderRequest)
	if err != nil {
		return exchange.Order{}, err
	}
	o := exchange.Order{
		Id:            convert.StrToUint64(r.Id),
		ClientOrderId: r.Text,
		Type:          r.Type, // limit
		Symbol:        r.CurrencyPair,
		Price:         decimal.RequireFromString(r.Price),
		Amount:        decimal.RequireFromString(r.Amount),
		Time:          time.Unix(convert.StrToInt64(r.CreateTime), 0),
		Status:        r.Status,
	}
	return o, err
}

// BuyFromTicker use Ticker's last price to place order.
//// may can't fill when big bull
func (g *SpotV4) buyFromTicker(ctx context.Context, symbol exchange.Symbol, clientOrderId string, total decimal.Decimal) (exchange.Order, error) {
	price, err := g.LastPrice(ctx, symbol.Symbol)
	if err != nil {
		return exchange.Order{}, err
	}
	price = price.Round(symbol.PricePrecision)
	amount := total.DivRound(price, symbol.AmountPrecision)
	return g.BuyLimit(ctx, symbol.Symbol, clientOrderId, price, amount)
}

// BuyFromOrderBook use order book and take the sell1
func (g *SpotV4) buyFromOrderBook(ctx context.Context, symbol exchange.Symbol, clientOrderId string, total decimal.Decimal) (exchange.Order, error) {
	ob, err := g.orderBook(ctx, symbol.Symbol)
	if err != nil {
		return exchange.Order{}, err
	}
	price, amount := takeAsks(ob.Asks, total, symbol.PricePrecision, symbol.AmountPrecision)
	return g.BuyLimit(ctx, symbol.Symbol, clientOrderId, price, amount)
}

// sellFromTicker use Ticker's last price to place order.
// may can't fill when big bear
func (g *SpotV4) sellFromTicker(ctx context.Context, symbol exchange.Symbol, clientOrderId string, amount decimal.Decimal) (exchange.Order, error) {
	price, err := g.LastPrice(ctx, symbol.Symbol)
	if err != nil {
		return exchange.Order{}, err
	}
	price = price.Round(symbol.PricePrecision)
	return g.SellLimit(ctx, symbol.Symbol, clientOrderId, price, amount)
}

// sellFromOrderBook use order book and take the buy1
func (g *SpotV4) sellFromOrderBook(ctx context.Context, symbol exchange.Symbol, clientOrderId string, amount decimal.Decimal) (exchange.Order, error) {
	ob, err := g.orderBook(ctx, symbol.Symbol)
	if err != nil {
		return exchange.Order{}, err
	}
	price := takeBids(ob.Bids, amount, symbol.PricePrecision, symbol.AmountPrecision)
	return g.SellLimit(ctx, symbol.Symbol, clientOrderId, price, amount)
}

func (g *SpotV4) SubscribeOrder(ctx context.Context, symbol, clientId string, responseHandler exchange.ResponseHandler) {
	g.SubOrder(ctx, symbol, clientId, responseHandler)
}

//func (g *SpotV4) UnsubscribeOrder(symbol, clientId string) {
//	g.UnsubOrder(symbol, clientId)
//}

func (g *SpotV4) SubscribeCandlestick(symbol, clientId string, period time.Duration, responseHandler exchange.ResponseHandler) {
	g.SubCandlestick(symbol, clientId, period, responseHandler)
}

func (g *SpotV4) UnsubscribeCandlestick(symbol, clientId string, period time.Duration) {
	g.UnsubCandlestick(symbol, clientId)
}

func (g *SpotV4) SubscribeCandlestickWithReq(symbol, clientId string, period time.Duration, responseHandler exchange.ResponseHandler) {
	panic("implement me")
}

func (g *SpotV4) UnsubscribeCandlestickWithReq(symbol, clientId string, period time.Duration) {
	panic("implement me")
}

//func (g *SpotV4) GetFee(symbol string) (fee exchange.Fee, err error) {
//	fee.Symbol = symbol
//	fee.BaseMaker = decimal.NewFromFloat(DefaultMaker)
//	fee.BaseTaker = decimal.NewFromFloat(DefaultTaker)
//	fee.ActualMaker = fee.BaseMaker
//	fee.ActualTaker = fee.BaseTaker
//	return
//}

//// Depth
//func (g *SpotV4) orderBooks() string {
//	var method string = "GET"
//	url := "/orderBooks"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}

// Depth of pair
func (g *SpotV4) orderBook(ctx context.Context, symbol string) (exchange.OrderBook, error) {
	ob, _, err := g.client.SpotApi.ListOrderBook(ctx, symbol, &gateapi.ListOrderBookOpts{})
	if err != nil {
		return exchange.OrderBook{}, err
	}
	return convertOrderBook(ob), err
}

// Trade History
//func (g *SpotV4) TradeHistory(params string) (string, error) {
//	url := "/TradeHistory/" + params
//	param := ""
//	data, err := g.httpDo(GET, url, param)
//	if err != nil {
//		return "", err
//	} else {
//		return string(data), err
//	}
//}

//// get deposit address
//func (g *SpotV4) depositAddress(currency string) string {
//	var method string = "POST"
//	url := "/private/depositAddress"
//	param := "currency=" + currency
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
//// get deposit withdrawal history
//func (g *SpotV4) depositsWithdrawals(start string, end string) string {
//	var method string = "POST"
//	url := "/private/depositsWithdrawals"
//	param := "start=" + start + "&end=" + end
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//

func (g *SpotV4) BuyStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

func (g *SpotV4) SellStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

// 获取我的24小时内成交记录
//func (g *SpotV4) MyTradeHistory(currencyPair string) (*MyTradeHistoryResult, error) {
//	method := "POST"
//	url := "/private/TradeHistory"
//	param := "orderNumber=&currencyPair=" + currencyPair
//	var result MyTradeHistoryResult
//	if err := g.request(method, url, param, &result); err != nil {
//		return nil, err
//	} else {
//		return &result, nil
//	}
//}

// Get my last 24h trades
//func (g *SpotV4) withdraw(currency string, amount string, address string) string {
//	var method string = "POST"
//	url := "/private/withdraw"
//	param := "currency=" + currency + "&amount=" + amount + "&address=" + address
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}

// always call ReqPing use context context with a timeout
func (g *SpotV4) ReqPing(ctx context.Context, id int64) (string, error) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	ch := make(chan string, 1)
	client.SetHandler(
		func() {
			g.Logger.Debug("successfully connected")
			client.Ping(id)
		},
		func(resp interface{}) {
			pong, ok := resp.(string)
			if ok {
				g.Logger.Debugf("handler got response: %s", pong)
				ch <- pong
			} else {
				g.Logger.Error("wrong response")
			}
		},
	)
	client.Connect(true)
	defer client.Close()

	for {
		select {
		case pong := <-ch:
			g.Logger.Debugf("response: %s", pong)
			return pong, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func (g *SpotV4) ReqTime(ctx context.Context, id int64) (int64, error) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	ch := make(chan int64, 1)
	client.SetHandler(
		func() {
			g.Logger.Debug("successfully connected")
			client.Time(id)
		},
		client.TimeHandler(func(resp interface{}) {
			t, ok := resp.(int64)
			if ok {
				g.Logger.Debugf("handler got response: %d", t)
				ch <- t
			} else {
				g.Logger.Error("wrong response")
			}
		}),
	)
	client.Connect(true)
	defer client.Close()

	for {
		select {
		case t := <-ch:
			g.Logger.Debugf("response timestamp: %d", t)
			return t, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}

func (g *SpotV4) ReqTicker(ctx context.Context, id int64, symbol string, period time.Duration) (hs.Ticker, error) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	ch := make(chan hs.Ticker, 1)
	client.SetHandler(
		func() {
			g.Logger.Debug("successfully connected")
			client.ReqTicker(id, symbol, int64(period.Seconds()))
		},
		client.ReqTickerHandler(func(resp interface{}) {
			t1, ok := resp.(ResponseWsTicker)
			if !ok {
				g.Logger.Error("wrong response")
				return
			}
			g.Logger.Debugf("handler got raw response: %d", t1)
			t2, _ := parseTicker(t1)
			ch <- t2
		}),
	)
	client.Connect(true)
	defer client.Close()

	for {
		select {
		case t := <-ch:
			g.Logger.Debugf("response timestamp: %d", t)
			return t, nil
		case <-ctx.Done():
			return hs.Ticker{}, ctx.Err()
		}
	}
}

func (g *SpotV4) SubTicker(id int64, symbol string, responseHandler exchange.ResponseHandler) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	client.SetHandler(
		func() {
			client.SubTicker(id, symbol)
		},
		client.SubTickerHandler(responseHandler),
	)
	client.Connect(true)
}

func (g *SpotV4) UnsubTicker(id int64, symbol string) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	client.UnsubTicker(id)
}

func (g *SpotV4) ReqCandlestick(ctx context.Context, symbol, clientId string, period time.Duration, from, to time.Time) (hs.Candle, error) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	ch := make(chan hs.Candle, 1)
	id := time.Now().Unix()
	client.SetHandler(
		func() {
			client.ReqCandle(id, symbol, from.Unix(), to.Unix(), int64(period.Seconds()))
		},
		client.ReqCandleHandler(func(resp interface{}) {
			r, ok := resp.(hs.Candle)
			if !ok {
				return
			}
			ch <- r
		}),
	)
	client.Connect(true)
	defer client.Close()

	for {
		select {
		case c := <-ch:
			return c, nil
		case <-ctx.Done():
			return hs.Candle{}, ctx.Err()
		}
	}
}

func (g *SpotV4) SubCandlestick(symbol, clientId string, period time.Duration,
	responseHandler exchange.ResponseHandler) {
	id := time.Now().Unix()
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	client.SetHandler(
		func() {
			client.SubCandle(id, symbol, int64(period.Seconds()))
		},
		client.SubCandleHandler(responseHandler),
	)
	client.Connect(true)
}

func (g *SpotV4) UnsubCandlestick(symbol, clientId string) {
	client := new(WebsocketClient).Init(g.wsHost, g.wsPath, g.Logger)
	id := time.Now().Unix()
	client.UnsubCandle(id)
}

func (g *SpotV4) ReqOrder(ctx context.Context, symbol, clientId string) (orders []exchange.Order, err error) {
	client := new(PrivateWebsocketClient).Init(g.wsHost, g.wsPath, g.Key, g.Secret, g.Logger)
	id := time.Now().Unix()
	ch := make(chan ResponseReqOrder, 10)
	done := make(chan int, 1)
	var offset uint64 = 0
	var limit uint64 = 10
	client.SetHandler(
		func() {
			client.ReqOrder(id, symbol, offset, limit)
		},
		client.ReqOrderHandler(func(resp interface{}) {
			if resp == nil {
				g.Logger.Debug("no open order")
				done <- 1
				return
			}
			r, ok := resp.(ResponseReqOrder)
			if !ok {
				g.Logger.Debug("response not ok")
				done <- 1
				return
			}
			ch <- r
		}),
	)
	client.Connect(true)
	defer client.Close()

	for {
		select {
		case <-done:
			return
		case batch := <-ch:
			g.Logger.Debugf("received batch data, offset %d, limit %d, total %d, len %d", batch.Offset, batch.Limit, batch.Total, len(batch.Records))
			if batch.Limit <= batch.Total {
				offset += limit
				client.ReqOrder(id, symbol, offset, limit)
			}
			if o, err := parseOrdersQuery(batch.Records); err == nil {
				g.Logger.Debugf("parsed %d orders", len(o))
				orders = append(orders, o...)
			} else {
				g.Logger.Errorf("got bad order: %s", err)
			}
			if batch.Offset+int64(len(batch.Records)) >= batch.Total {
				return
			}
		case <-ctx.Done():
			return orders, ctx.Err()
		}
	}
}

// TODO: 应该改造为使用context，当context取消时，自动退出订阅
func (g *SpotV4) SubOrder(ctx context.Context, symbol, clientId string, responseHandler exchange.ResponseHandler) {
	id := time.Now().Unix()
	client := new(PrivateWebsocketClient).Init(g.wsHost, g.wsPath, g.Key, g.Secret, g.Logger)
	client.SetHandler(
		func() {
			client.SubOrder(id, []string{symbol})
		},
		client.SubOrderHandler(responseHandler),
	)
	client.Connect(true)
	<-ctx.Done()
	client.UnsubOrder(id, []string{symbol})
}

//func (g *SpotV4) UnsubOrder(symbol, clientId string) {
//	client := new(PrivateWebsocketClient).Init(g.wsHost, g.wsPath, g.Key, g.Secret, g.Logger)
//	id := time.Now().Unix()
//	client.UnsubOrder(id, []string{symbol})
//}

func (g *SpotV4) ReqBalance(ctx context.Context, currencies []string) {

}

func (g *SpotV4) SubBalance(currencies []string) {

}

func (g *SpotV4) UnsubBalance(currencies []string) {

}
