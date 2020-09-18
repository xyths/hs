package huobi

import (
	"context"
	"errors"
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/client"
	"github.com/huobirdcenter/huobi_golang/pkg/client/accountwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/marketwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/orderwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
	"github.com/huobirdcenter/huobi_golang/pkg/model/account"
	"github.com/huobirdcenter/huobi_golang/pkg/model/auth"
	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
	"github.com/huobirdcenter/huobi_golang/pkg/model/order"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
	"github.com/xyths/hs/exchange"
	"github.com/xyths/hs/logger"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	Name        = "huobi"
	DefaultHost = "api.huobi.me"
)

type Client struct {
	Label     string
	AccessKey string
	SecretKey string
	Host      string

	SpotAccountId int64
}

func New(label, accessKey, secretKey, host string) (*Client, error) {
	c := &Client{
		Label:     label,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
	if host != "" {
		c.Host = host
	} else {
		c.Host = DefaultHost
	}
	accountId, err := c.GetSpotAccountId()
	if err == nil {
		c.SpotAccountId = accountId
	} else {
		return nil, err
	}
	return c, nil
}

func (c *Client) GetTimestamp() (int, error) {
	hb := new(client.CommonClient).Init(c.Host)
	return hb.GetTimestamp()
}

func (c *Client) AllSymbols() (s []exchange.Symbol, err error) {
	hb := new(client.CommonClient).Init(c.Host)
	symbols, err := hb.GetSymbols()
	if err != nil {
		return
	}
	s = make([]exchange.Symbol, len(symbols))
	for i, a := range symbols {
		s[i] = exchange.Symbol{
			Symbol:          a.Symbol,
			BaseCurrency:    a.BaseCurrency,
			QuoteCurrency:   a.QuoteCurrency,
			AmountPrecision: int32(a.AmountPrecision),
			PricePrecision:  int32(a.PricePrecision),
			MinAmount:       a.MinOrderAmt,
			MinTotal:        a.MinOrderValue,
		}
	}
	return
}

func (c *Client) GetSymbol(symbol string) (s exchange.Symbol, err error) {
	hb := new(client.CommonClient).Init(c.Host)
	symbols, err := hb.GetSymbols()
	if err != nil {
		return
	}
	for _, a := range symbols {
		if a.Symbol == symbol {
			s.Symbol = a.Symbol
			s.BaseCurrency = a.BaseCurrency
			s.QuoteCurrency = a.QuoteCurrency
			s.AmountPrecision = int32(a.AmountPrecision)
			s.PricePrecision = int32(a.PricePrecision)
			s.MinAmount = a.MinOrderAmt
			s.MinTotal = a.MinOrderValue
			break
		}
	}
	return
}

func (c *Client) GetAccountInfo() ([]account.AccountInfo, error) {
	hb := new(client.AccountClient).Init(c.AccessKey, c.SecretKey, c.Host)
	return hb.GetAccountInfo()
}

func (c *Client) GetSpotAccountId() (int64, error) {
	accounts, err := c.GetAccountInfo()
	if err != nil {
		return 0, err
	}
	for _, a := range accounts {
		if a.Type == "spot" {
			return a.Id, nil
		}
	}
	return 0, nil
}

func (c *Client) LastPrice(symbol string) (decimal.Decimal, error) {
	hb := new(client.MarketClient).Init(c.Host)

	optionalRequest := market.GetCandlestickOptionalRequest{Period: market.MIN1, Size: 1}
	candlesticks, err := hb.GetCandlestick(symbol, optionalRequest)
	if err != nil {
		log.Println(err)
		return decimal.NewFromFloat(0), err
	}
	for _, candlestick := range candlesticks {
		log.Printf("1min candlestick: OHLC[%s, %s, %s, %s]",
			candlestick.Open, candlestick.High, candlestick.Low, candlestick.Close)
		return candlestick.Close, nil
	}

	return decimal.NewFromFloat(0), nil
}

func (c *Client) SpotAvailableBalance() (map[string]decimal.Decimal, error) {
	hb := new(client.AccountClient).Init(c.AccessKey, c.SecretKey, c.Host)
	accountBalance, err := hb.GetAccountBalance(fmt.Sprintf("%d", c.SpotAccountId))
	if err != nil {
		return nil, err
	}
	balance := make(map[string]decimal.Decimal)
	for _, b := range accountBalance.List {
		nb, err := decimal.NewFromString(b.Balance)
		if err != nil {
			log.Printf("error when parse balance: %s", err)
			continue
		}
		if nb.IsZero() {
			continue
		}
		if ob, ok := balance[b.Currency]; ok {
			balance[b.Currency] = ob.Add(nb)
		} else {
			balance[b.Currency] = nb
		}
	}
	return balance, nil
}

type CandleSlice []hs.Candle

func (cs CandleSlice) Len() int {
	return len(cs)
}
func (cs CandleSlice) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}
func (cs CandleSlice) Less(i, j int) bool {
	return cs[i].Timestamp[0] < cs[j].Timestamp[0]
}

func (c *Client) CandleBySize(symbol string, period time.Duration, size int) (hs.Candle, error) {
	hb := new(client.MarketClient).Init(c.Host)
	optionalRequest := market.GetCandlestickOptionalRequest{Period: getPeriodString(period), Size: size}
	candlesticks, err := hb.GetCandlestick(symbol, optionalRequest)
	if err != nil {
		return hs.Candle{}, err
	}
	l := len(candlesticks)
	candle := hs.NewCandle(l)
	candle.Timestamp = make([]int64, l)
	candle.Open = make([]float64, l)
	candle.High = make([]float64, l)
	candle.Low = make([]float64, l)
	candle.Close = make([]float64, l)
	candle.Volume = make([]float64, l)
	for i := l - 1; i >= 0; i-- {
		candle.Timestamp[l-1-i] = candlesticks[i].Id
		candle.Open[l-1-i], _ = candlesticks[i].Open.Float64()
		candle.High[l-1-i], _ = candlesticks[i].High.Float64()
		candle.Low[l-1-i], _ = candlesticks[i].Low.Float64()
		candle.Close[l-1-i], _ = candlesticks[i].Close.Float64()
		candle.Volume[l-1-i], _ = candlesticks[i].Vol.Float64()
	}

	return candle, nil
}

func (c *Client) CandleFrom(symbol, clientId string, period time.Duration, from, to time.Time) (hs.Candle, error) {
	timestamps := c.splitTimestamp(period, from, to)
	if len(timestamps) <= 1 {
		return hs.Candle{}, errors.New("'from' need before 'to'")
	}
	hb := new(marketwebsocketclient.CandlestickWebSocketClient).Init(c.Host)
	ch := make(chan hs.Candle, len(timestamps)-1)
	candles := hs.NewCandle(CandlestickReqMaxLength * (len(timestamps) - 1))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var cs CandleSlice
		for i := 0; i < len(timestamps)-1; i++ {
			candle := <-ch
			cs = append(cs, candle)
		}
		sort.Sort(cs)
		for _, c := range cs {
			candles.Add(c)
		}
	}()
	periodStr := getPeriodString(period)
	hb.SetHandler(
		// Connected handler
		func() {
			for i := 1; i < len(timestamps); i++ {
				hb.Request(symbol, periodStr, timestamps[i-1], timestamps[i], clientId)
				wg.Add(1)
				time.Sleep(time.Second / 10)
			}
		},
		func(resp interface{}) {
			defer wg.Done()
			candlestickResponse, ok := resp.(market.SubscribeCandlestickResponse)
			if ok {
				if &candlestickResponse != nil {
					if candlestickResponse.Tick != nil {
						t := candlestickResponse.Tick
						logger.Sugar.Infof("Candlestick update, id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
							t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)
					}

					if candlestickResponse.Data != nil {
						candle := hs.NewCandle(CandlestickReqMaxLength)
						for i := 0; i < len(candlestickResponse.Data); i++ {
							tick := candlestickResponse.Data[i]
							ticker := hs.Ticker{
								Timestamp: tick.Id,
							}
							ticker.Open, _ = tick.Open.Float64()
							ticker.High, _ = tick.High.Float64()
							ticker.Low, _ = tick.Low.Float64()
							ticker.Close, _ = tick.Close.Float64()
							ticker.Volume, _ = tick.Vol.Float64()
							candle.Append(ticker)
						}
						ch <- candle
						return
					}
				}
			} else {
				logger.Sugar.Warn("Unknown response: %v", resp)
			}
			ch <- hs.Candle{}
			return
		})

	hb.Connect(true)
	defer hb.UnSubscribe(symbol, periodStr, clientId)

	wg.Wait()

	return candles, nil
}

func (c *Client) GetOrderById(orderId uint64, symbol string) (exchange.Order, error) {
	hb := new(client.OrderClient).Init(c.AccessKey, c.SecretKey, c.Host)
	r, err := hb.GetOrderById(fmt.Sprint(orderId))
	if err != nil {
		return exchange.Order{}, err
	}
	d := r.Data
	o := exchange.Order{
		Id:            uint64(d.Id),
		ClientOrderId: d.ClientOrderId,
		Type:          d.Type,
		Symbol:        d.Symbol,
		InitialPrice:  decimal.RequireFromString(d.Price),
		InitialAmount: decimal.RequireFromString(d.Amount),
		Timestamp:     d.CreatedAt,
		Status:        d.State,
		FilledAmount:  decimal.RequireFromString(d.FilledAmount),
	}
	return o, nil
}

func (c *Client) PlaceOrder(request *order.PlaceOrderRequest) (uint64, error) {
	hb := new(client.OrderClient).Init(c.AccessKey, c.SecretKey, c.Host)
	resp, err := hb.PlaceOrder(request)
	if err != nil {
		return 0, err
	}
	switch resp.Status {
	case "ok":
		log.Printf("Place order successfully, order id: %s, clientOrderId: %s\n", resp.Data, request.ClientOrderId)
		return convert.StrToUint64(resp.Data), nil
	case "error":
		log.Printf("Place order error: %s\n", resp.ErrorMessage)
		if resp.ErrorCode == "account-frozen-balance-insufficient-error" {
			return 0, nil
		}
		return 0, errors.New(resp.ErrorMessage)
	}
	return 0, errors.New("unknown status")
}

func (c *Client) SpotLimitOrder(orderType, symbol, clientOrderId string, price, amount decimal.Decimal) (uint64, error) {
	request := order.PlaceOrderRequest{
		AccountId:     fmt.Sprintf("%d", c.SpotAccountId),
		Type:          orderType,
		Source:        "spot-api",
		Symbol:        symbol,
		Price:         price.String(),
		Amount:        amount.String(),
		ClientOrderId: clientOrderId,
	}
	return c.PlaceOrder(&request)
}

func (c *Client) SpotMarketOrder(orderType, symbol, clientOrderId string, total decimal.Decimal) (uint64, error) {
	request := order.PlaceOrderRequest{
		AccountId:     fmt.Sprintf("%d", c.SpotAccountId),
		Type:          orderType,
		Source:        "spot-api",
		Symbol:        symbol,
		Amount:        total.String(),
		ClientOrderId: clientOrderId,
	}
	return c.PlaceOrder(&request)
}

func (c *Client) SpotStopLimitOrder(orderType, symbol, clientOrderId, operator string, price, amount, stopPrice decimal.Decimal) (uint64, error) {
	request := order.PlaceOrderRequest{
		AccountId:     fmt.Sprintf("%d", c.SpotAccountId),
		Type:          orderType,
		Source:        "spot-api",
		Symbol:        symbol,
		Price:         price.String(),
		Amount:        amount.String(),
		ClientOrderId: clientOrderId,
		StopPrice:     stopPrice.String(),
		Operator:      operator,
	}
	return c.PlaceOrder(&request)
}

func (c *Client) BuyLimit(symbol, clientOrderId string, price, amount decimal.Decimal) (orderId uint64, err error) {
	return c.SpotLimitOrder(OrderTypeBuyLimit, symbol, clientOrderId, price, amount)
}

func (c *Client) SellLimit(symbol, clientOrderId string, price, amount decimal.Decimal) (orderId uint64, err error) {
	return c.SpotLimitOrder(OrderTypeSellLimit, symbol, clientOrderId, price, amount)
}

func (c *Client) BuyMarket(symbol, clientOrderId string, total decimal.Decimal) (orderId uint64, err error) {
	return c.SpotMarketOrder(OrderTypeBuyMarket, symbol, clientOrderId, total)
}

func (c *Client) SellMarket(symbol, clientOrderId string, total decimal.Decimal) (orderId uint64, err error) {
	return c.SpotMarketOrder(OrderTypeSellMarket, symbol, clientOrderId, total)
}

func (c *Client) BuyStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return c.SpotStopLimitOrder(OrderTypeBuyStopLimit, symbol, clientOrderId, "gte", price, amount, stopPrice)
}

func (c *Client) SellStopLimit(symbol, clientOrderId string, price, amount, stopPrice decimal.Decimal) (orderId uint64, err error) {
	return c.SpotStopLimitOrder(OrderTypeSellStopLimit, symbol, clientOrderId, "lte", price, amount, stopPrice)
}

func (c *Client) CancelOrder(symbol string, orderId uint64) error {
	hb := new(client.OrderClient).Init(c.AccessKey, c.SecretKey, c.Host)
	resp, err := hb.CancelOrderById(fmt.Sprintf("%d", orderId))
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}
	errorCode, err := strconv.Atoi(resp.ErrorCode)
	if err != nil {
		return nil
	}
	if errorCode == 0 {
		return nil
	} else {
		return errors.New(resp.ErrorMessage)
	}
}

func (c *Client) SubscribeLast24hCandlestick(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) error {
	hb := new(marketwebsocketclient.Last24hCandlestickWebSocketClient).Init(c.Host)
	hb.SetHandler(
		// Connected handler
		func() {
			hb.Subscribe(symbol, clientId)
		},
		responseHandler)

	hb.Connect(true)

	select {
	case <-ctx.Done():
		hb.UnSubscribe(symbol, clientId)
		log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
	}
	return nil
}

func (c *Client) SubscribeCandlestick(ctx context.Context, symbol, clientId string, period time.Duration,
	responseHandler exchange.ResponseHandler) {
	periodStr := getPeriodString(period)
	hb := new(marketwebsocketclient.CandlestickWebSocketClient).Init(c.Host)
	hb.SetHandler(
		// Connected handler
		func() {
			hb.Subscribe(symbol, periodStr, clientId)
		},
		websocketclientbase.ResponseHandler(responseHandler),
	)

	hb.Connect(true)

	<-ctx.Done()

	hb.UnSubscribe(symbol, periodStr, clientId)
	log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
}

const CandlestickReqMaxLength = 300

func (c *Client) SubscribeCandlestickWithReq(ctx context.Context, symbol, clientId string, period time.Duration,
	responseHandler exchange.ResponseHandler) {
	hb := new(marketwebsocketclient.CandlestickWebSocketClient).Init(c.Host)
	now := time.Now()
	periodStr := getPeriodString(period)
	start := now.Add(-CandlestickReqMaxLength * period)
	hb.SetHandler(
		// Connected handler
		func() {
			hb.Request(symbol, periodStr, start.Unix(), now.Unix(), clientId)
			hb.Subscribe(symbol, periodStr, clientId)
		},
		websocketclientbase.ResponseHandler(responseHandler))

	hb.Connect(true)

	<-ctx.Done()

	hb.UnSubscribe(symbol, periodStr, clientId)
	log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
}

func (c *Client) SubscribeOrder(ctx context.Context, symbol, clientId string,
	responseHandler exchange.ResponseHandler) {
	hb := new(orderwebsocketclient.SubscribeOrderWebSocketV2Client).Init(c.AccessKey, c.SecretKey, c.Host)

	hb.SetHandler(
		// Connected handler
		func(resp *auth.WebSocketV2AuthenticationResponse) {
			if resp.IsSuccess() {
				// Subscribe if authentication passed
				hb.Subscribe(symbol, clientId)
			} else {
				log.Fatalf("Authentication error, code: %d, message:%s", resp.Code, resp.Message)
			}
		},
		websocketclientbase.ResponseHandler(responseHandler))

	hb.Connect(true)

	<-ctx.Done()

	hb.UnSubscribe(symbol, clientId)
	log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
}

func (c *Client) SubscribeAccountUpdate(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) error {
	hb := new(accountwebsocketclient.SubscribeAccountWebSocketV2Client).Init(c.AccessKey, c.SecretKey, c.Host)

	hb.SetHandler(
		// Connected handler
		func(resp *auth.WebSocketV2AuthenticationResponse) {
			if resp.IsSuccess() {
				// Subscribe if authentication passed
				hb.Subscribe("1", clientId)
			} else {
				applogger.Error("Authentication error, code: %d, message:%s", resp.Code, resp.Message)
			}
		},
		responseHandler)

	hb.Connect(true)

	<-ctx.Done()

	hb.UnSubscribe("1", clientId)
	log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)

	return nil
}

func (c *Client) SubscribeTradeClear(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) {
	hb := new(orderwebsocketclient.SubscribeTradeClearWebSocketV2Client).Init(c.AccessKey, c.SecretKey, c.Host)

	hb.SetHandler(
		// Connected handler
		func(resp *auth.WebSocketV2AuthenticationResponse) {
			if resp.IsSuccess() {
				// Subscribe if authentication passed
				hb.Subscribe(symbol, clientId)
			} else {
				applogger.Error("Authentication error, code: %d, message:%s", resp.Code, resp.Message)
			}
		},
		responseHandler)

	hb.Connect(true)

	<-ctx.Done()

	hb.UnSubscribe(symbol, clientId)
	log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
}

func (c Client) splitTimestamp(period time.Duration, from, to time.Time) (timestamps []int64) {
	//var d time.Duration
	//switch period {
	//case market.MIN1:
	//	d = time.Minute
	//case market.MIN5:
	//	d = time.Minute * 5
	//case market.MIN15:
	//	d = time.Minute * 15
	//case market.MIN30:
	//	d = time.Minute * 30
	//case market.MIN60:
	//	d = time.Hour
	//case market.HOUR4:
	//	d = time.Hour * 4
	//case market.DAY1:
	//	d = time.Hour * 24
	//case market.MON1:
	//	d = time.Hour * 24 * 30
	//case market.WEEK1:
	//	d = time.Hour * 24 * 7
	//case market.YEAR1:
	//	d = time.Hour * 24 * 365
	//default:
	//	d = time.Hour * 24
	//}

	for t := from; t.Before(to); t = t.Add(period * CandlestickReqMaxLength) {
		timestamps = append(timestamps, t.Unix())
	}
	timestamps = append(timestamps, to.Unix())

	return
}

func getPeriodString(period time.Duration) (periodStr string) {
	switch period {
	case exchange.MIN1:
		periodStr = market.MIN1
	case exchange.MIN5:
		periodStr = market.MIN5
	case exchange.MIN15:
		periodStr = market.MIN15
	case exchange.MIN30:
		periodStr = market.MIN30
	case exchange.HOUR1:
		periodStr = market.MIN60
	case exchange.HOUR4:
		periodStr = market.HOUR4
	case exchange.DAY1:
		periodStr = market.DAY1
	case exchange.MON1:
		periodStr = market.MON1
	case exchange.WEEK1:
		periodStr = market.WEEK1
	case exchange.YEAR1:
		periodStr = market.YEAR1
	default:
		logger.Sugar.Fatalf("bad period")
	}
	return
}
