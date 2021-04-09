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
			Time:          time.Unix(int64(r.CTime), 0),
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
	o.Time = time.Unix(int64(r.CTime), 0)
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
// use in V4 API, spot and futures
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
		Time:          time.Unix(convert.StrToInt64(o.CreateTime), 0),
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

func convertTrade(t gateapi.Trade) exchange.Trade {
	return exchange.Trade{
		Id:      convert.StrToUint64(t.Id),
		OrderId: convert.StrToUint64(t.OrderId),
		//Symbol:        t.CurrencyPair,
		Side:        t.Side,
		Role:        t.Role,
		Price:       decimal.RequireFromString(t.Price),
		Amount:      decimal.RequireFromString(t.Amount),
		Time:        time.Unix(convert.StrToInt64(t.CreateTime), 0),
		FeeCurrency: t.FeeCurrency,
		FeeAmount:   decimal.RequireFromString(t.Fee),
	}
}

func convertContract(c gateapi.Contract) exchange.Contract {
	return exchange.Contract{
		Name:              c.Name,
		Type:              c.Type,
		QuoteMultiplier:   c.QuantoMultiplier,
		LeverageMax:       c.LeverageMax,
		LeverageMin:       c.LeverageMin,
		MaintenanceRate:   c.MaintenanceRate,
		MarkType:          c.MarkType,
		MarkPrice:         c.MarkPrice,
		IndexPrice:        c.IndexPrice,
		LastPrice:         c.LastPrice,
		MakerFeeRate:      c.MakerFeeRate,
		TakerFeeRate:      c.TakerFeeRate,
		OrderPriceRound:   c.OrderPriceRound,
		MarkPriceRound:    c.MarkPriceRound,
		FundingRate:       c.FundingRate,
		FundingInterval:   c.FundingInterval,
		FundingNextApply:  c.FundingNextApply,
		RiskLimitBase:     c.RiskLimitBase,
		RiskLimitStep:     c.RiskLimitStep,
		RiskLimitMax:      c.RiskLimitMax,
		OrderSizeMax:      c.OrderSizeMax,
		OrderSizeMin:      c.OrderSizeMin,
		OrderPriceDeviate: c.OrderPriceDeviate,
		RefDiscountRate:   c.RefDiscountRate,
		RefRebateRate:     c.RefRebateRate,
		OrderbookId:       c.OrderbookId,
		TradeId:           c.TradeId,
		TradeSize:         c.TradeSize,
		PositionSize:      c.PositionSize,
		ConfigChangeTime:  c.ConfigChangeTime,
		InDelisting:       c.InDelisting,
		OrdersLimit:       c.OrdersLimit,
	}
}

func convertFutureOrderBook(fob gateapi.FuturesOrderBook) exchange.FuturesOrderbook {
	cob := exchange.FuturesOrderbook{}
	for _, ask := range fob.Asks {
		cob.Asks = append(cob.Asks, exchange.FuturesQuote{Price: convert.StrToFloat64(ask.P), Amount: ask.S})
	}
	for _, bid := range fob.Bids {
		cob.Bids = append(cob.Bids, exchange.FuturesQuote{Price: convert.StrToFloat64(bid.P), Amount: bid.S})
	}
	return cob
}

func convertCommonTrade(t gateapi.FuturesTrade) exchange.FuturesTrade {
	return exchange.FuturesTrade{
		Id:         uint64(t.Id),
		CreateTime: time.Unix(int64(t.CreateTime), 0),
		Contract:   t.Contract,
		Size:       t.Size,
		Price:      decimal.RequireFromString(t.Price),
	}
}

func convertCommonLiquidation(liq gateapi.FuturesLiquidate) exchange.FuturesLiquidation {
	return exchange.FuturesLiquidation{
		Time:       time.Unix(liq.Time, 0),
		Contract:   liq.Contract,
		Size:       liq.Size,
		OrderPrice: decimal.RequireFromString(liq.OrderPrice),
		FillPrice:  decimal.RequireFromString(liq.FillPrice),
		Left:       liq.Left,
	}
}

func convertBalance(account gateapi.FuturesAccount) exchange.FuturesBalance {
	return exchange.FuturesBalance{
		Currency:       account.Currency,
		Available:      decimal.RequireFromString(account.Available),
		PositionMargin: decimal.RequireFromString(account.PositionMargin),
		OrderMargin:    decimal.RequireFromString(account.OrderMargin),
		UnrealisedPnl:  decimal.RequireFromString(account.UnrealisedPnl),
		Total:          decimal.RequireFromString(account.Total),
		Dual:           account.InDualMode,
	}
}

func convertPosition(p gateapi.Position) exchange.Position {
	var closeOrder *exchange.PositionCloseOrder
	if p.CloseOrder.Id > 0 {
		closeOrder = &exchange.PositionCloseOrder{
			Id:    p.CloseOrder.Id,
			Price: decimal.RequireFromString(p.CloseOrder.Price),
			IsLiq: p.CloseOrder.IsLiq,
		}
	}
	return exchange.Position{
		User:            p.User,
		Contract:        p.Contract,
		Size:            p.Size,
		Leverage:        int(convert.StrToInt64(p.Leverage)),
		RiskLimit:       int(convert.StrToInt64(p.RiskLimit)),
		MaxLeverage:     int(convert.StrToInt64(p.LeverageMax)),
		MaintenanceRate: convert.StrToFloat64(p.MaintenanceRate),
		Value:           decimal.RequireFromString(p.Value),
		Margin:          decimal.RequireFromString(p.Margin),
		EntryPrice:      decimal.RequireFromString(p.EntryPrice),
		LiqPrice:        decimal.RequireFromString(p.LiqPrice),
		MarkPrice:       decimal.RequireFromString(p.MarkPrice),
		UnrealisedPnl:   decimal.RequireFromString(p.UnrealisedPnl),
		RealisedPnl:     decimal.RequireFromString(p.RealisedPnl),
		HistoryPnl:      decimal.RequireFromString(p.HistoryPnl),
		LastClosePnl:    decimal.RequireFromString(p.LastClosePnl),
		AdlRanking:      p.AdlRanking,
		PendingOrders:   p.PendingOrders,
		CloseOrder:      closeOrder,
		Mode:            p.Mode,
	}
}

func convertDualPosition(positions []gateapi.Position) []exchange.Position {
	var rets []exchange.Position
	for _, p := range positions {
		rets = append(rets, convertPosition(p))
	}
	return rets
}
