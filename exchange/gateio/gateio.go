package gateio

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
	"github.com/xyths/hs/exchange"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultHost = "gateio.life"
)

type GateIO struct {
	Key    string
	Secret string

	publicBaseUrl  string
	privateBaseUrl string
}

func New(key, secret, host string) *GateIO {
	g := &GateIO{Key: key, Secret: secret}
	if host == "" {
		host = DefaultHost
	}
	g.publicBaseUrl = "https://data." + host + "/api2/1"
	g.privateBaseUrl = "https://api." + host + "/api2/1"
	return g
}

const (
	GET  = "GET"
	POST = "POST"
)

// all support pairs
func (g *GateIO) GetPairs() (string, error) {
	url := "/pairs"
	param := ""
	if ret, err := g.httpDo(GET, url, param); err != nil {
		return "", err
	} else {
		return string(ret), err
	}
}
func (g *GateIO) AllSymbols() (s []exchange.Symbol, err error) {
	panic("Not implemented")
}
func (g *GateIO) GetSymbol(symbol string) (s exchange.Symbol, err error) {
	s.Symbol = symbol
	switch symbol {
	case BTC_USDT:
		s.BaseCurrency = BTC
		s.QuoteCurrency = USDT
	case SERO_USDT:
		s.BaseCurrency = SERO
		s.QuoteCurrency = USDT
	case ETH:
		s.BaseCurrency = ETH
		s.QuoteCurrency = USDT
	}
	s.PricePrecision = PricePrecision[symbol]
	s.AmountPrecision = AmountPrecision[symbol]
	s.MinAmount = decimal.NewFromFloat(MinAmount[symbol])
	s.MinTotal = decimal.NewFromFloat(DefaultMinTotal)
	return
}
func (g *GateIO) GetFee(symbol string) (fee exchange.Fee, err error) {
	fee.Symbol = symbol
	fee.BaseMaker = decimal.NewFromFloat(DefaultMaker)
	fee.BaseTaker = decimal.NewFromFloat(DefaultTaker)
	fee.ActualMaker = fee.BaseMaker
	fee.ActualTaker = fee.BaseTaker
	return
}

// Market Info
//func (g *GateIO) marketinfo() string {
//	var method string = "GET"
//	url :=  "/marketinfo"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
//// Market Details
//func (g *GateIO) marketlist() string {
//	var method string = "GET"
//	url := "/marketlist"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
//// tickers
//func (g *GateIO) tickers() string {
//	var method string = "GET"
//	url := "/tickers"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
// ticker
func (g *GateIO) Ticker(currencyPair string) (*exchange.Ticker, error) {
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

func (g *GateIO) LastPrice(symbol string) (decimal.Decimal, error) {
	ticker, err := g.Ticker(symbol)
	if err != nil {
		return decimal.Zero, err
	}
	return ticker.Last, nil
}

//// Depth
//func (g *GateIO) orderBooks() string {
//	var method string = "GET"
//	url := "/orderBooks"
//	param := ""
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}

// Depth of pair
func (g *GateIO) OrderBook(symbol string) (ResponseOrderBook, error) {
	var method string = "GET"
	url := "/orderBook/" + symbol
	param := ""

	var result ResponseOrderBook
	err := g.request(method, url, param, &result)
	return result, err
}

// 获取Candle
func (g *GateIO) CandleBySize(symbol string, period time.Duration, size int) (candle hs.Candle, err error) {
	groupSec := int(period.Seconds())
	rangeHour := size / int(time.Hour/period)
	return g.GetCandle(symbol, groupSec, rangeHour)
}

func (g *GateIO) CandleFrom(symbol, clientId string, period time.Duration, from, to time.Time) (hs.Candle, error) {
	return hs.Candle{}, nil
}

// 获取Candle
func (g *GateIO) GetCandle(symbol string, groupSec, rangeHour int) (candles hs.Candle, err error) {
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
func (g *GateIO) TradeHistory(params string) (string, error) {
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
func (g *GateIO) SpotBalance() (map[string]decimal.Decimal, error) {
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
func (g *GateIO) SpotAvailableBalance() (map[string]decimal.Decimal, error) {
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

//// get deposit address
//func (g *GateIO) depositAddress(currency string) string {
//	var method string = "POST"
//	url := "/private/depositAddress"
//	param := "currency=" + currency
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//
//// get deposit withdrawal history
//func (g *GateIO) depositsWithdrawals(start string, end string) string {
//	var method string = "POST"
//	url := "/private/depositsWithdrawals"
//	param := "start=" + start + "&end=" + end
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}
//

// 订单类型("gtc"：普通订单（默认）；“ioc”：立即执行否则取消订单（Immediate-Or-Cancel，IOC）；"poc":被动委托（只挂单，不吃单）（Pending-Or-Cancelled，POC）)
// Place order buy
func (g *GateIO) BuyLimit(symbol, text string, price, amount decimal.Decimal) (orderId uint64, err error) {
	resp, err := g.BuyOrder(symbol, price, amount, OrderTypeGTC, text)
	if err != nil {
		return 0, err
	}
	if resp.Result == "false" || resp.OrderNumber == 0 {
		return 0, errors.New(resp.Message)
	}
	return resp.OrderNumber, nil
}

func (g *GateIO) BuyOrder(symbol string, price, amount decimal.Decimal, orderType, text string) (resp ResponseOrder, err error) {
	url := "/private/buy"
	param := fmt.Sprintf("currencyPair=%s&rate=%s&amount=%s&orderType=%s&text=t-%s", symbol, price, amount, orderType, text)
	err = g.request(POST, url, param, &resp)
	return
}

// Place order sell
func (g *GateIO) SellLimit(symbol, text string, price, amount decimal.Decimal) (orderId uint64, err error) {
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

func (g *GateIO) BuyMarket(symbol, clientOrderId string, amount decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

func (g *GateIO) SellMarket(symbol, clientOrderId string, amount decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

func (g *GateIO) BuyStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

func (g *GateIO) SellStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return 0, nil
}

// Cancel order
func (g *GateIO) CancelOrder(symbol string, orderNumber uint64) error {
	url := "/private/cancelOrder"
	param := fmt.Sprintf("currencyPair=%s&orderNumber=%d", symbol, orderNumber)
	var res ResponseCancel
	err := g.request(POST, url, param, &res)
	//ok = res.Result
	return err
}

// Cancel all orders
func (g *GateIO) CancelAllOrders(types string, currencyPair string) (res ResponseCancel, err error) {
	url := "/private/cancelAllOrders"
	param := "type=" + types + "&currencyPair=" + currencyPair
	err = g.request(POST, url, param, &res)
	return
}

func (g *GateIO) GetOrderById(orderId uint64, symbol string) (order exchange.Order, err error) {
	return g.GetOrder(orderId, symbol)
}

// Get order status
func (g *GateIO) GetOrder(orderNumber uint64, currencyPair string) (order exchange.Order, err error) {
	url := "/private/getOrder"
	param := fmt.Sprintf("orderNumber=%d&currencyPair=%s", orderNumber, currencyPair)
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
	order.Timestamp = o.Timestamp

	return
}

func (g *GateIO) IsOrderClose(symbol string, orderId uint64) (order exchange.Order, closed bool) {
	o, err := g.GetOrder(orderId, symbol)
	if err != nil {
		return o, false
	}
	if o.Status == "closed" {
		return o, true
	}
	return o, false
}

// Get my open order list
func (g *GateIO) OpenOrders() (res ResponseOpenOrders, err error) {
	url := "/private/openOrders"
	param := ""
	err = g.request(POST, url, param, &res)
	return
}

// 获取我的24小时内成交记录
func (g *GateIO) MyTradeHistory(currencyPair string) (*MyTradeHistoryResult, error) {
	method := "POST"
	url := "/private/TradeHistory"
	param := "orderNumber=&currencyPair=" + currencyPair
	var result MyTradeHistoryResult
	if err := g.request(method, url, param, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

// Get my last 24h trades
//func (g *GateIO) withdraw(currency string, amount string, address string) string {
//	var method string = "POST"
//	url := "/private/withdraw"
//	param := "currency=" + currency + "&amount=" + amount + "&address=" + address
//	var ret string = g.httpDo(method, url, param)
//	return ret
//}

func (g *GateIO) getSign(params string) string {
	key := []byte(g.Secret)
	mac := hmac.New(sha512.New, key)
	mac.Write([]byte(params))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

/**
*  http request
 */
func (g *GateIO) httpDo(method string, url string, param string) ([]byte, error) {
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

func (g *GateIO) request(method string, url string, param string, result interface{}) error {
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
