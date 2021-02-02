package gateio

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gateio/gateapi-go/v5"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
	"github.com/xyths/hs/exchange"
	"strings"
	"time"
)

func (g GateIO) FormatSymbol(base, quote string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(base), strings.ToLower(quote))
}

func parseTicker(data ResponseWsTicker) (ticker hs.Ticker, err error) {
	// no ticker.Timestamp
	ticker.Open = convert.StrToFloat64(data.Open)
	ticker.High = convert.StrToFloat64(data.High)
	ticker.Low = convert.StrToFloat64(data.Low)
	ticker.Close = convert.StrToFloat64(data.Close)
	ticker.Volume = convert.StrToFloat64(data.BaseVolume)
	return
}

func parseCandle(data []interface{}) (hs.Candle, error) {
	c := hs.NewCandle(len(data))
	var err error

	for _, d := range data {
		raw, ok := d.([]interface{})
		if !ok {
			err = errors.New("ticker format error")
			break
		}
		if len(raw) != 8 {
			err = errors.New("ticker not 8 item")
			break
		}
		t := hs.Ticker{}
		timestamp, ok := raw[0].(float64)
		if !ok {
			err = errors.New("timestamp format error")
			break
		}
		t.Timestamp = int64(timestamp)
		prices := make([]float64, 5)
		for i := 1; i < 6; i++ {
			prices[i-1] = convert.StrToFloat64(raw[i].(string))
		}
		t.Open = prices[0]
		t.Close = prices[1]
		t.High = prices[2]
		t.Low = prices[3]
		t.Volume = prices[4]
		c.Append(t)
	}
	return c, err
}

func parseTickers(data []interface{}) ([]hs.Ticker, error) {
	var tickers []hs.Ticker
	var err error

	for _, d := range data {
		raw, ok := d.([]interface{})
		if !ok {
			err = errors.New("ticker format error")
			break
		}
		if len(raw) != 8 {
			err = errors.New("ticker not 8 item")
			break
		}
		t := hs.Ticker{}
		timestamp, ok := raw[0].(float64)
		if !ok {
			err = errors.New("timestamp format error")
			break
		}
		t.Timestamp = int64(timestamp)
		prices := make([]float64, 5)
		for i := 1; i < 6; i++ {
			prices[i-1] = convert.StrToFloat64(raw[i].(string))
		}
		t.Open = prices[0]
		t.Close = prices[1]
		t.High = prices[2]
		t.Low = prices[3]
		t.Volume = prices[4]
		tickers = append(tickers, t)
	}
	return tickers, err
}

// parseOrdersQuery parse the records in response of order.query
func parseOrdersQuery(records []WsOrderRecord) (orders []exchange.Order, err error) {
	for _, r := range records {
		orders = append(orders, exchange.Order{
			Id:            r.Id,
			ClientOrderId: r.Text,
			Type:          parseOrderType(r.Type),
			Symbol:        r.Market,
			Price:         decimal.RequireFromString(r.Price),
			Amount:        decimal.RequireFromString(r.Amount),
			Timestamp:     int64(r.CTime),
			//Status:
			FilledAmount: decimal.RequireFromString(r.FilledAmount),
		})
	}
	return
}

func parseOrderType(gateType int) string {
	switch gateType {
	case 0:
		return "unknown"
	default:
		return "unknown"
	}
}

// parseOrderUpdate parse order.update message's params field
func parseOrderUpdate(params []interface{}) (exchange.Order, error) {
	o := exchange.Order{}

	if len(params) != 2 {
		return o, errors.New("order.update should have 2 params")
	}
	event, ok := params[0].(int)
	if !ok {
		return o, errors.New("bad event in order.update message")
	}
	r := WsOrderRecord{}
	data, err := json.Marshal(params[1])
	if err != nil {
		return o, err
	}
	if err1 := json.Unmarshal(data, &r); err1 != nil {
		return o, err
	}
	o.Id = r.Id
	o.ClientOrderId = r.Text
	o.Type = parseOrderType(r.Type)
	o.Symbol = r.Market
	o.Price = decimal.RequireFromString(r.Price)
	o.Amount = decimal.RequireFromString(r.Amount)
	o.Timestamp = int64(r.CTime)
	o.FilledAmount = decimal.RequireFromString(r.FilledAmount)
	// event type,Integer, 1: PUT, 2: UPDATE, 3: FINISH
	switch event {
	case 1:
		o.Status = "open"
	case 2:
		o.Status = "filled"
	case 3:
		o.Status = "finish"
	}
	return o, nil
}

func cutSymbol(symbol string) (base, quote string) {
	tokens := strings.Split(symbol, "_")
	if len(tokens) == 2 {
		base = strings.ToUpper(tokens[0])
		quote = strings.ToUpper(tokens[1])
	}
	return
}

func takeAsks(asks []Quote, total decimal.Decimal, pricePrecision, amountPrecision int32) (price, amount decimal.Decimal) {
	left := total
	for i := len(asks) - 1; i >= 0; i-- {
		price = decimal.NewFromFloat(asks[i][0]).Round(pricePrecision)
		a := decimal.NewFromFloat(asks[i][1]).Round(amountPrecision)
		taken := left.DivRound(price, amountPrecision)
		if taken.GreaterThan(a) {
			taken = a
			amount = amount.Add(taken)
			left = left.Sub(amount.Mul(price))
		} else {
			amount = amount.Add(taken)
			break
		}
	}
	// fix amount, this will reduce the amount
	amount = total.DivRound(price, amountPrecision)
	return
}

func takeBids(bids []Quote, amount decimal.Decimal, pricePrecision, amountPrecision int32) (price decimal.Decimal) {
	left := amount
	for i := 0; i < len(bids); i++ {
		price = decimal.NewFromFloat(bids[i][0]).Round(pricePrecision)
		a := decimal.NewFromFloat(bids[i][1]).Round(amountPrecision)
		taken := left.DivRound(price, amountPrecision)
		if taken.GreaterThan(a) {
			taken = a
			left = left.Sub(taken)
		} else {
			break
		}
	}
	return
}

// getInterval convert duration to gate candle period
// use in V4 API
func getInterval(period time.Duration) string {
	switch period {
	case time.Second * 10:
		return "10s"
	case time.Minute:
		return "1m"
	case time.Minute * 5:
		return "5m"
	case time.Minute * 15:
		return "15m"
	case time.Minute * 30:
		return "30m"
	case time.Hour:
		return "1h"
	case time.Hour * 4:
		return "4h"
	case time.Hour * 8:
		return "8h"
	case time.Hour * 24:
		return "1d"
	case time.Hour * 24 * 7:
		return "7d"
	default:
		return "1d"
	}
}

func convertOrder(o gateapi.Order) exchange.Order {
	return exchange.Order{
		Id:            convert.StrToUint64(o.Id),
		ClientOrderId: o.Text,
		Type:          o.Type, // limit
		Symbol:        o.CurrencyPair,
		Price:         decimal.RequireFromString(o.Price),
		Amount:        decimal.RequireFromString(o.Amount),
		Timestamp:     convert.StrToInt64(o.CreateTime),
		Status:        o.Status,
	}
}

func convertTicker(t gateapi.Ticker) exchange.Ticker {
	return exchange.Ticker{
		Last:          decimal.RequireFromString(t.Last),
		LowestAsk:     decimal.RequireFromString(t.LowestAsk),
		HighestBid:    decimal.RequireFromString(t.HighestBid),
		PercentChange: decimal.RequireFromString(t.ChangePercentage),
		BaseVolume:    decimal.RequireFromString(t.BaseVolume),
		QuoteVolume:   decimal.RequireFromString(t.QuoteVolume),
		High24hr:      decimal.RequireFromString(t.High24h),
		Low24hr:       decimal.RequireFromString(t.Low24h),
	}
}

func convertOrderBook(ob gateapi.OrderBook) exchange.OrderBook {
	cob := exchange.OrderBook{Id: int(ob.Id)}
	for _, ask := range ob.Asks {
		if len(ask) != 2 {
			continue
		}
		cob.Asks = append(cob.Asks, exchange.Quote{convert.StrToFloat64(ask[0]), convert.StrToFloat64(ask[1])})
	}
	for _, bid := range ob.Bids {
		if len(bid) != 2 {
			continue
		}
		cob.Bids = append(cob.Bids, exchange.Quote{convert.StrToFloat64(bid[0]), convert.StrToFloat64(bid[1])})
	}
	return cob
}
