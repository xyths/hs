// V4 Futures RESTful API

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
	"strconv"
	"time"
)

type Futures struct {
	Key    string
	Secret string
	client *gateapi.APIClient
	wsHost string
	wsPath string

	Logger *zap.SugaredLogger
}

func NewFutures(key, secret, host string, logger *zap.SugaredLogger) *Futures {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	return &Futures{Key: key, Secret: secret, client: client, wsHost: host, wsPath: "/v4", Logger: logger}
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

func (f *Futures) ListContracts(ctx context.Context, settle string) ([]exchange.Contract, error) {
	rawContracts, _, err := f.client.FuturesApi.ListFuturesContracts(ctx, settle)
	if err != nil {
		return nil, err
	}
	var contracts []exchange.Contract
	for _, c := range rawContracts {
		contract := convertContract(c)
		contracts = append(contracts, contract)
	}
	return contracts, err
}

func (f *Futures) GetContract(ctx context.Context, settle, contract string) (exchange.Contract, error) {
	rawContract, _, err := f.client.FuturesApi.GetFuturesContract(ctx, settle, contract)
	if err != nil {
		return exchange.Contract{}, err
	}
	return convertContract(rawContract), err
}

func (f *Futures) Orderbook(ctx context.Context, settle, contract string, limit int, interval float32) (exchange.FuturesOrderbook, error) {
	opts := gateapi.ListFuturesOrderBookOpts{Limit: optional.NewInt32(int32(limit))}
	if interval != 0 {
		str := fmt.Sprintf("%f", interval)
		opts.Interval = optional.NewString(str)
	}
	raw, _, err := f.client.FuturesApi.ListFuturesOrderBook(ctx, settle, contract, &opts)
	if err != nil {
		return exchange.FuturesOrderbook{}, err
	}
	return convertFutureOrderBook(raw), err
}

func (f *Futures) ListTrades(ctx context.Context, settle string, contract string, limit int, from, to int64) ([]exchange.FuturesTrade, error) {
	opts := gateapi.ListFuturesTradesOpts{}
	if limit > 0 {
		opts.Limit = optional.NewInt32(int32(limit))
	}
	if from > 0 {
		opts.From = optional.NewInt64(from)
	}
	if to > 0 {
		opts.To = optional.NewInt64(to)
	}
	raw, _, err := f.client.FuturesApi.ListFuturesTrades(ctx, settle, contract, &opts)
	if err != nil {
		return nil, err
	}
	var trades []exchange.FuturesTrade
	for _, t := range raw {
		trades = append(trades, convertCommonTrade(t))
	}
	return trades, err
}

// from, to is unix timestamp in seconds
func (f *Futures) Candle(ctx context.Context, settle string, contract string, from, to int64, limit int, interval time.Duration) (hs.Candle, error) {
	opts := gateapi.ListFuturesCandlesticksOpts{Interval: optional.NewString(getInterval(interval))}
	if from > 0 {
		opts.From = optional.NewInt64(from)
	}
	if to > 0 {
		opts.To = optional.NewInt64(to)
	}
	if limit > 0 {
		opts.Limit = optional.NewInt32(int32(limit))
	}
	rawCandle, _, err := f.client.FuturesApi.ListFuturesCandlesticks(ctx, settle, contract, &opts)
	if err != nil {
		return hs.Candle{}, err
	}
	candle := hs.NewCandle(len(rawCandle))
	for _, c := range rawCandle {
		candle.Append(hs.Ticker{
			Timestamp: int64(c.T), // unix timestamp in seconds
			Open:      convert.StrToFloat64(c.O),
			High:      convert.StrToFloat64(c.H),
			Low:       convert.StrToFloat64(c.L),
			Close:     convert.StrToFloat64(c.C),
			Volume:    float64(c.V), // raw data is int64
		})
	}
	return candle, err
}

func (f *Futures) CommonLiquidation(ctx context.Context, settle, contract string, from, to int64, limit int) ([]exchange.FuturesLiquidation, error) {
	opts := gateapi.ListLiquidatedOrdersOpts{Contract: optional.NewString(contract)}
	if from > 0 {
		opts.From = optional.NewInt64(from)
	}
	if to > 0 {
		opts.To = optional.NewInt64(to)
	}
	if limit > 0 {
		opts.Limit = optional.NewInt32(int32(limit))
	}
	rawOrders, _, err := f.client.FuturesApi.ListLiquidatedOrders(ctx, settle, &opts)
	if err != nil {
		return nil, err
	}
	var liquidations []exchange.FuturesLiquidation
	for _, o := range rawOrders {
		liquidations = append(liquidations, convertCommonLiquidation(o))
	}
	return liquidations, err
}

func (f *Futures) ListFuturesAccounts(ctx context.Context, settle string) (exchange.FuturesBalance, error) {
	raw, _, err := f.client.FuturesApi.ListFuturesAccounts(ctx, settle)
	if err != nil {
		return exchange.FuturesBalance{}, err
	}
	return convertBalance(raw), nil
}

// 设置持仓模式
func (f *Futures) SetDualMode(ctx context.Context, settle string, newDualMode bool) (exchange.FuturesBalance, error) {
	raw, _, err := f.client.FuturesApi.SetDualMode(ctx, settle, newDualMode)
	if err != nil {
		return exchange.FuturesBalance{}, err
	}
	return convertBalance(raw), err
}

func (f *Futures) ListPositions(ctx context.Context, settle string) ([]exchange.Position, error) {
	rawList, _, err := f.client.FuturesApi.ListPositions(ctx, settle)
	if err != nil {
		return nil, err
	}
	var positions []exchange.Position
	for _, p := range rawList {
		positions = append(positions, convertPosition(p))
	}
	return positions, err
}

func (f *Futures) GetPosition(ctx context.Context, settle, contract string) (exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.GetPosition(ctx, settle, contract)
	if err != nil {
		return exchange.Position{}, err
	}
	return convertPosition(raw), err
}

func (f *Futures) AddMargin(ctx context.Context, settle, contract string, margin decimal.Decimal) (exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.UpdatePositionMargin(ctx, settle, contract, margin.String())
	if err != nil {
		return exchange.Position{}, err
	}
	return convertPosition(raw), err
}

// 更新头寸杠杆
func (f *Futures) UpdateLeverage(ctx context.Context, settle, contract string, newLeverage int) (exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.UpdatePositionLeverage(ctx, settle, contract, strconv.Itoa(newLeverage))
	if err != nil {
		return exchange.Position{}, err
	}
	return convertPosition(raw), err
}

// 更新头寸风险限额
func (f *Futures) UpdateRiskLimit(ctx context.Context, settle, contract string, newRiskLimit int) (exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.UpdatePositionRiskLimit(ctx, settle, contract, strconv.Itoa(newRiskLimit))
	if err != nil {
		return exchange.Position{}, err
	}
	return convertPosition(raw), err
}

func (f *Futures) GetDualPosition(ctx context.Context, settle string, contract string) ([]exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.GetDualModePosition(ctx, settle, contract)
	if err != nil {
		return nil, err
	}
	return convertDualPosition(raw), err
}

func (f *Futures) AddDualMargin(ctx context.Context, settle, contract string, margin decimal.Decimal) ([]exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.UpdateDualModePositionMargin(ctx, settle, contract, margin.String())
	if err != nil {
		return nil, err
	}
	return convertDualPosition(raw), err
}

// 更新双仓模式下的头寸杠杆
func (f *Futures) UpdateDualLeverage(ctx context.Context, settle, contract string, newLeverage int) ([]exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.UpdateDualModePositionLeverage(ctx, settle, contract, strconv.Itoa(newLeverage))
	if err != nil {
		return nil, err
	}
	return convertDualPosition(raw), err
}

// 更新双仓模式下的头寸风险限额
func (f *Futures) UpdateDualRiskLimit(ctx context.Context, settle, contract string, newRiskLimit int) ([]exchange.Position, error) {
	raw, _, err := f.client.FuturesApi.UpdateDualModePositionRiskLimit(ctx, settle, contract, strconv.Itoa(newRiskLimit))
	if err != nil {
		return nil, err
	}
	return convertDualPosition(raw), err
}