package gateio

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
	"github.com/xyths/hs/exchange"
	"strings"
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
