package gateio

import (
	"errors"
	"fmt"
	"github.com/xyths/hs"
	"github.com/xyths/hs/convert"
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
