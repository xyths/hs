// V2版本的Gate接口，仅支持现货，仅支持RESTful
// 支持websocket比较有限，仅复制了一些函数，未经反复测试
// 此V2会逐步废弃

package gateio

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
	"github.com/xyths/hs/exchange"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type V2 struct {
	Key    string
	Secret string

	publicBaseUrl  string
	privateBaseUrl string

	host   string
	wsPath string

	Logger *zap.SugaredLogger
}

func NewV2(key, secret, host string, logger *zap.SugaredLogger) *V2 {
	g := &V2{Key: key, Secret: secret, wsPath: WsPathV4, Logger: logger}
	if host == "" {
		host = DefaultHost
	}
	g.host = host
	g.publicBaseUrl = "https://data." + host + "/api2/1"
	g.privateBaseUrl = "https://api." + host + "/api2/1"
	return g
}

// all support pairs
func (g *V2) GetPairs() (pairs []string, err error) {
	var method string = "GET"
	url := "/pairs"
	param := ""
	err = g.request(method, url, param, &pairs)
	return
}

func (g *V2) AllSymbols(ctx context.Context) (symbols []exchange.Symbol, err error) {
	infos, err := g.MarketInfo()
	if err != nil {
		return
	}
	for _, stupidMap := range infos.Pairs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			for symbol, info := range stupidMap {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					base, quote := cutSymbol(symbol)
					symbols = append(symbols, exchange.Symbol{
						Symbol:              symbol,
						Disabled:            info.TradeDisabled != 0 || info.BuyDisabled != 0 || info.SellDisabled != 0,
						BaseCurrency:        base,
						QuoteCurrency:       quote,
						PricePrecision:      info.PricePrecision,
						AmountPrecision:     info.AmountPrecision,
						LimitOrderMinAmount: decimal.NewFromFloat(info.MinAmount),
						MinTotal:            decimal.NewFromFloat(info.MinAmountB),
					})
				}
			}
		}
	}
	return
}

func (g *V2) GetSymbol(ctx context.Context, symbol string) (s exchange.Symbol, err error) {
	all, err := g.AllSymbols(ctx)
	if err != nil {
		return
	}
	for _, s_ := range all {
		select {
		case <-ctx.Done():
			return s, ctx.Err()
		default:
			if s_.Symbol == symbol {
				s = s_
				break
			}
		}
	}
	return
}

func (g *V2) GetFee(symbol string) (fee exchange.Fee, err error) {
	fee.Symbol = symbol
	fee.BaseMaker = decimal.NewFromFloat(DefaultMaker)
	fee.BaseTaker = decimal.NewFromFloat(DefaultTaker)
	fee.ActualMaker = fee.BaseMaker
	fee.ActualTaker = fee.BaseTaker
	return
}

// Market Info
func (g *V2) MarketInfo() (res ResponseMarketInfo, err error) {
	var method string = "GET"
	url := "/marketinfo"
	param := ""
	err = g.request(method, url, param, &res)
	return
}

//// Market Details
//func (g *V2) marketlist() string {
//	var method string = "GET"
//	url := "/marketlist"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
//// tickers
//func (g *V2) tickers() string {
//	var method string = "GET"
//	url := "/tickers"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
// ticker
func (g *V2) Ticker(currencyPair string) (*exchange.Ticker, error) {
	url := "/ticker" + "/" + currencyPair
	param := ""
	var t ResponseTicker
	if err := g.request(GET, url, param, &t); err != nil {
		return nil, err
	}
	ticker := &exchange.Ticker{
		Last:          decimal.RequireFromString(t.Last),
		LowestAsk:     decimal.RequireFromString(t.LowestAsk),
		HighestBid:    decimal.RequireFromString(t.HighestBid),
		PercentChange: decimal.RequireFromString(t.PercentChange),
		BaseVolume:    decimal.RequireFromString(t.BaseVolume),
		QuoteVolume:   decimal.RequireFromString(t.QuoteVolume),
		High24hr:      decimal.RequireFromString(t.High24hr),
		Low24hr:       decimal.RequireFromString(t.Low24hr),
	}
	return ticker, nil
}

func (g *V2) LastPrice(symbol string) (decimal.Decimal, error) {
	ticker, err := g.Ticker(symbol)
	if err != nil {
		return decimal.Zero, err
	}
	return ticker.Last, nil
}

func (g *V2) Last24hVolume(symbol string) (decimal.Decimal, error) {
	ticker, err := g.Ticker(symbol)
	if err != nil {
		return decimal.Zero, err
	}
	return ticker.BaseVolume, nil
}

//// Depth
//func (g *V2) orderBooks() string {
//	var method string = "GET"
//	url := "/orderBooks"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}

// Depth of pair
func (g *V2) OrderBook(symbol string) (ResponseOrderBook, error) {
	var method string = "GET"
	url := "/orderBook/" + symbol
	param := ""

	var result ResponseOrderBook
	err := g.request(method, url, param, &result)
	return result, err
}

// 获取Candle
func (g *V2) CandleBySize(symbol string, period time.Duration, size int) (candle hs.Candle, err error) {
	groupSec := int(period.Seconds())
	rangeHour := int(int64(size) * int64(period) / int64(time.Hour))
	return g.GetCandle(symbol, groupSec, rangeHour)
}

func (g *V2) CandleFrom(symbol, clientId string, period time.Duration, from, to time.Time) (hs.Candle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	return g.ReqCandlestick(ctx, symbol, clientId, period, from, to)
}

// 获取Candle
func (g *V2) GetCandle(symbol string, groupSec, rangeHour int) (candles hs.Candle, err error) {
	url := fmt.Sprintf("/candlestick2/%s?group_sec=%d&range_hour=%d", symbol, groupSec, rangeHour)
	param := ""

	var result ResponseCandles
	err = g.request(GET, url, param, &result)
	if err != nil {
		return candles, err
	}
	candles = hs.NewCandle(len(result.Data))
	for i := 0; i < len(result.Data); i++ {
		c := result.Data[i]
		candles.Append(hs.Ticker{
			Timestamp: int64(c[0] / 1000), // covert ms to s
			Volume:    c[1],
			Close:     c[2],
			High:      c[3],
			Low:       c[4],
			Open:      c[5],
		})
	}
	return
}

// Trade History
func (g *V2) TradeHistory(params string) (string, error) {
	url := "/TradeHistory/" + params
	param := ""
	data, err := g.httpDo(GET, url, param)
	if err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Get account fund balances
func (g *V2) SpotBalance() (map[string]decimal.Decimal, error) {
	url := "/private/balances"
	param := ""
	data, err := g.httpDo(POST, url, param)
	if err != nil {
		return nil, err
	}
	var result ResponseBalances
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	balance := make(map[string]decimal.Decimal)
	for k, v := range result.Available {
		b := decimal.RequireFromString(v)
		if b.IsZero() {
			continue
		}
		if ob, ok := balance[k]; ok {
			balance[k] = ob.Add(b)
		} else {
			balance[k] = b
		}
	}
	for k, v := range result.Locked {
		b := decimal.RequireFromString(v)
		if b.IsZero() {
			continue
		}
		if ob, ok := balance[k]; ok {
			balance[k] = ob.Add(b)
		} else {
			balance[k] = b
		}
	}
	return balance, nil
}

// Get account fund balances
func (g *V2) SpotAvailableBalance() (map[string]decimal.Decimal, error) {
	url := "/private/balances"
	param := ""
	data, err := g.httpDo(POST, url, param)
	if err != nil {
		return nil, err
	}
	var result ResponseBalances
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	balance := make(map[string]decimal.Decimal)
	for k, v := range result.Available {
		b := decimal.RequireFromString(v)
		if b.IsZero() {
			continue
		}
		if ob, ok := balance[k]; ok {
			balance[k] = ob.Add(b)
		} else {
			balance[k] = b
		}
	}
	return balance, nil
}

// Get spot balances in details
func (g *V2) SpotBalanceDetail() (map[string]exchange.Balance, error) {
	url := "/private/balances"
	param := ""
	data, err := g.httpDo(POST, url, param)
	if err != nil {
		return nil, err
	}
	var result ResponseBalances
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	balance := make(map[string]exchange.Balance)
	for k, v := range result.Available {
		b := decimal.RequireFromString(v)
		if b.IsZero() {
			continue
		}
		if _, ok := balance[k]; ok {
			balance[k].Available.Add(b)
		} else {
			balance[k] = exchange.Balance{
				Currency:  k,
				Available: b,
			}
		}
	}
	for k, v := range result.Locked {
		b := decimal.RequireFromString(v)
		if b.IsZero() {
			continue
		}
		if _, ok := balance[k]; ok {
			balance[k].Locked.Add(b)
		} else {
			balance[k] = exchange.Balance{
				Currency: k,
				Locked:   b,
			}
		}
	}
	return balance, nil
}

//// get deposit address
//func (g *V2) depositAddress(currency string) string {
//	var method string = "POST"
//	url := "/private/depositAddress"
//	param := "currency=" + currency
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
//// get deposit withdrawal history
//func (g *V2) depositsWithdrawals(start string, end string) string {
//	var method string = "POST"
//	url := "/private/depositsWithdrawals"
//	param := "start=" + start + "&end=" + end
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//

// 订单类型("gtc"：普通订单（默认）；“ioc”：立即执行否则取消订单（Immediate-Or-Cancel，IOC）；"poc":被动委托（只挂单，不吃单）（Pending-Or-Cancelled，POC）)
// Place order buy
func (g *V2) BuyLimit(symbol, text string, price, amount decimal.Decimal) (orderId uint64, err error) {
	resp, err := g.BuyOrder(symbol, price, amount, OrderTypeGTC, text)
	if err != nil {
		return 0, err
	}
	if resp.Result == "false" || resp.OrderNumber == 0 {
		return 0, errors.New(resp.Message)
	}
	return resp.OrderNumber, nil
}

func (g *V2) BuyOrder(symbol string, price, amount decimal.Decimal, orderType, text string) (resp ResponseOrder, err error) {
	url := "/private/buy"
	param := fmt.Sprintf("currencyPair=%s&rate=%s&amount=%s&orderType=%s&text=t-%s", symbol, price, amount, orderType, text)
	err = g.request(POST, url, param, &resp)
	return
}

// Place order sell
func (g *V2) SellLimit(symbol, text string, price, amount decimal.Decimal) (orderId uint64, err error) {
	url := "/private/sell"
	// 价格精度：5，数量精度：3
	param := fmt.Sprintf("currencyPair=%s&rate=%s&amount=%s&orderType=%s&text=t-%s", symbol, price, amount, OrderTypeGTC, text)
	var res ResponseOrder
	err = g.request(POST, url, param, &res)
	if err != nil {
		return 0, err
	}
	if res.Result == "false" || res.OrderNumber == 0 {
		return 0, errors.New(res.Message)
	}
	return res.OrderNumber, nil
}

func (g *V2) BuyMarket(symbol exchange.Symbol, clientOrderId string, total decimal.Decimal) (orderId uint64, err error) {
	price, err := g.LastPrice(symbol.Symbol)
	if err != nil {
		return
	}
	price = price.Round(symbol.PricePrecision)
	amount := total.DivRound(price, symbol.AmountPrecision)
	return g.BuyLimit(symbol.Symbol, clientOrderId, price, amount)
}

func (g *V2) SellMarket(symbol exchange.Symbol, clientOrderId string, amount decimal.Decimal) (orderId uint64, err error) {
	price, err := g.LastPrice(symbol.Symbol)
	if err != nil {
		return
	}
	price = price.Round(symbol.PricePrecision)
	return g.SellLimit(symbol.Symbol, clientOrderId, price, amount)
}

func (g *V2) BuyStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

func (g *V2) SellStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

// Cancel order
func (g *V2) CancelOrder(symbol string, orderNumber uint64) error {
	url := "/private/cancelOrder"
	param := fmt.Sprintf("currencyPair=%s&orderNumber=%d", symbol, orderNumber)
	var res ResponseCancel
	err := g.request(POST, url, param, &res)
	//ok = res.Result
	return err
}

// Cancel all orders
func (g *V2) CancelAllOrders(types string, currencyPair string) (res ResponseCancel, err error) {
	url := "/private/cancelAllOrders"
	param := "type=" + types + "&currencyPair=" + currencyPair
	err = g.request(POST, url, param, &res)
	return
}

func (g *V2) GetOrderById(orderId uint64, symbol string) (order exchange.Order, err error) {
	return g.GetOrder(orderId, symbol)
}

// Get order as string, just for test gate's getOrder interface, only use in gateio_test.go
func (g *V2) GetOrderString(orderId uint64, symbol string) (string, error) {
	url := "/private/getOrder"
	param := fmt.Sprintf("orderNumber=%d&currencyPair=%s", orderId, symbol)
	data, err := g.httpDo(POST, url, param)
	if err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Get order status
func (g *V2) GetOrder(orderId uint64, symbol string) (order exchange.Order, err error) {
	url := "/private/getOrder"
	param := fmt.Sprintf("orderNumber=%d&currencyPair=%s", orderId, symbol)
	var res ResponseGetOrder
	err = g.request(POST, url, param, &res)
	if err != nil {
		return
	}
	if res.Result != "true" || res.Message != "Success" {
		log.Printf("request not success: %#v", res)
		return order, errors.New(res.Message)
	}
	o := &res.Order
	order.Id = convert.StrToUint64(o.OrderNumber)
	order.ClientOrderId = o.Text
	order.Symbol = o.CurrencyPair
	order.Type = o.Type
	// 下单价格
	order.Price = decimal.RequireFromString(o.InitialRate)
	order.Amount = decimal.RequireFromString(o.InitialAmount)
	order.Status = o.Status
	order.FilledPrice = decimal.RequireFromString(o.Rate)
	order.FilledAmount = decimal.RequireFromString(o.FilledAmount)
	//order.FeePercentage = o.FeePercentage
	//order.FeeValue = decimal.RequireFromString(o.FeeValue)
	order.Time = time.Unix(o.Timestamp, 0)

	return
}

func (g *V2) IsFullFilled(symbol string, orderId uint64) (order exchange.Order, filled bool, err error) {
	order, err = g.GetOrder(orderId, symbol)
	if err != nil {
		return
	}
	filled = order.Status == "closed"
	return
}

// Get my open order list
func (g *V2) OpenOrders() ([]exchange.Order, error) {
	url := "/private/openOrders"
	param := ""
	var res ResponseOpenOrders
	err := g.request(POST, url, param, &res)
	if err != nil {
		return nil, err
	}
	var orders []exchange.Order
	for _, raw := range res.Orders {
		o := exchange.Order{
			Id:           raw.OrderNumber,
			Type:         raw.Type,
			Symbol:       raw.CurrencyPair,
			Price:        decimal.RequireFromString(raw.InitialRate),
			Amount:       decimal.RequireFromString(raw.InitialAmount),
			Time:         time.Unix(raw.Timestamp, 0),
			Status:       raw.Status,
			FilledPrice:  decimal.RequireFromString(raw.FilledRate),
			FilledAmount: decimal.RequireFromString(raw.FilledAmount),
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// 获取我的24小时内成交记录
func (g *V2) MyTradeHistory(symbol string, orderId uint64) ([]exchange.Trade, error) {
	method := "POST"
	url := "/private/TradeHistory"
	order := ""
	if orderId != 0 {
		order = fmt.Sprintf("%d", orderId)
	}
	param := fmt.Sprintf("orderNumber=%s&currencyPair=%s", order, symbol)
	var result MyTradeHistoryResult
	if err := g.request(method, url, param, &result); err != nil {
		return nil, err
	}
	var trades []exchange.Trade
	for _, r := range result.Trades {
		t := exchange.Trade{
			Id:      r.TradeId,
			OrderId: r.OrderNumber,
			Symbol:  r.Pair,
			Type:    r.Type,
			Price:   decimal.RequireFromString(r.Rate),
			Amount:  decimal.RequireFromString(r.Amount),
			Time:    time.Unix(r.TimeUnix, 0),
		}
		trades = append(trades, t)
	}
	return trades, nil
}

// Get my last 24h trades
//func (g *V2) withdraw(currency string, amount string, address string) string {
//	var method string = "POST"
//	url := "/private/withdraw"
//	param := "currency=" + currency + "&amount=" + amount + "&address=" + address
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}

func (g *V2) getSign(params string) string {
	key := []byte(g.Secret)
	mac := hmac.New(sha512.New, key)
	mac.Write([]byte(params))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

/**
*  http request
 */
func (g *V2) httpDo(method string, url string, param string) ([]byte, error) {
	client := &http.Client{}
	if method == GET {
		url = g.publicBaseUrl + url
	} else if method == POST {
		url = g.privateBaseUrl + url
	} else {
		return nil, errors.New("unknown method")
	}

	req, err := http.NewRequest(method, url, strings.NewReader(param))
	if err != nil {
		return nil, err
	}
	sign := g.getSign(param)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("key", g.Key)
	req.Header.Set("sign", sign)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}

	return body, nil
}

func (g *V2) request(method string, url string, param string, result interface{}) error {
	client := &http.Client{}
	if method == GET {
		url = g.publicBaseUrl + url
	} else if method == POST {
		url = g.privateBaseUrl + url
	} else {
		return errors.New("unsupported method")
	}

	req, err := http.NewRequest(method, url, strings.NewReader(param))
	if err != nil {
		return err
	}
	sign := g.getSign(param)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("key", g.Key)
	req.Header.Set("sign", sign)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Printf("error: %s", err)
		return err
	}
	if err = json.Unmarshal(data, result); err != nil {
		log.Printf("raw response: %s", string(data))
	}
	return err
}

// always call ReqPing use context context with a timeout
func (g *V2) ReqPing(ctx context.Context, id int64) (string, error) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
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

func (g *V2) ReqTime(ctx context.Context, id int64) (int64, error) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
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

func (g *V2) ReqTicker(ctx context.Context, id int64, symbol string, period time.Duration) (hs.Ticker, error) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
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

func (g *V2) SubTicker(id int64, symbol string, responseHandler exchange.ResponseHandler) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
	client.SetHandler(
		func() {
			client.SubTicker(id, symbol)
		},
		client.SubTickerHandler(responseHandler),
	)
	client.Connect(true)
}

func (g *V2) UnsubTicker(id int64, symbol string) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
	client.UnsubTicker(id)
}

func (g *V2) ReqCandlestick(ctx context.Context, symbol, clientId string, period time.Duration, from, to time.Time) (hs.Candle, error) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
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

func (g *V2) SubCandlestick(symbol, clientId string, period time.Duration,
	responseHandler exchange.ResponseHandler) {
	id := time.Now().Unix()
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
	client.SetHandler(
		func() {
			client.SubCandle(id, symbol, int64(period.Seconds()))
		},
		client.SubCandleHandler(responseHandler),
	)
	client.Connect(true)
}

func (g *V2) UnsubCandlestick(symbol, clientId string) {
	client := new(WebsocketClient).Init(g.host, g.wsPath, g.Logger)
	id := time.Now().Unix()
	client.UnsubCandle(id)
}

func (g *V2) ReqOrder(ctx context.Context, symbol, clientId string) (orders []exchange.Order, err error) {
	client := new(PrivateWebsocketClient).Init(g.host, g.wsPath, g.Key, g.Secret, g.Logger)
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

func (g *V2) SubOrder(symbol, clientId string, responseHandler exchange.ResponseHandler) {
	id := time.Now().Unix()
	client := new(PrivateWebsocketClient).Init(g.host, g.wsPath, g.Key, g.Secret, g.Logger)
	client.SetHandler(
		func() {
			client.SubOrder(id, []string{symbol})
		},
		client.SubOrderHandler(responseHandler),
	)
	client.Connect(true)
}

func (g *V2) UnsubOrder(symbol, clientId string) {
	client := new(PrivateWebsocketClient).Init(g.host, g.wsPath, g.Key, g.Secret, g.Logger)
	id := time.Now().Unix()
	client.UnsubOrder(id, []string{symbol})
}

func (g *V2) ReqBalance(ctx context.Context, currencies []string) {

}

func (g *V2) SubBalance(currencies []string) {

}

func (g *V2) UnsubBalance(currencies []string) {

}
