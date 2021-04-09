package gateio

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/xyths/hs"
	"testing"
	"time"
)

func TestFutures_ListContracts(t *testing.T) {
	settles := []string{"btc", "usdt"}
	ctx := context.TODO()
	for _, settle := range settles {
		contracts, err := futures.ListContracts(ctx, settle)
		require.NoError(t, err)
		for i, c := range contracts {
			b, err1 := json.MarshalIndent(c, "", "  ")
			require.NoError(t, err1)
			t.Logf("[%d] contract info\n%s", i, string(b))
		}
	}
}

func TestFutures_GetContract(t *testing.T) {
	settles := []string{"btc", "usdt"}
	ctx := context.TODO()
	for _, settle := range settles {
		contracts, err := futures.ListContracts(ctx, settle)
		require.NoError(t, err)
		for i, c := range contracts {
			contract, err1 := futures.GetContract(ctx, settle, c.Name)
			require.NoError(t, err1)
			b, err1 := json.MarshalIndent(contract, "", "  ")
			require.NoError(t, err1)
			t.Logf("[%d] contract info\n%s", i, string(b))
		}
	}
}

func TestFutures_Orderbook(t *testing.T) {
	var tests = []struct {
		Settle   string
		Contract string
		Interval float32
		Limit    int
	}{
		{"usdt", "BTC_USDT", 0.0, 5},
		{"usdt", "BTC_USDT", 0.1, 5},
		{"usdt", "BTC_USDT", 0.01, 5},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		orderbook, err := futures.Orderbook(ctx, tt.Settle, tt.Contract, tt.Limit, tt.Interval)
		require.NoError(t, err)
		t.Logf("orderbook %s %s %f %d \n%s", tt.Settle, tt.Contract, tt.Interval, tt.Limit, orderbook.String())
	}
}

func TestFutures_ListTrades(t *testing.T) {
	trades, err := futures.ListTrades(context.TODO(), "usdt", "BTC_USDT", 10, 0, 0)
	require.NoError(t, err)
	for i, trade := range trades {
		b, err1 := json.MarshalIndent(trade, "", "  ")
		require.NoError(t, err1)
		t.Logf("[%d] %s", i, string(b))
	}
}

func TestFutures_Candle(t *testing.T) {
	ctx := context.TODO()
	print := func(t *testing.T, candle hs.Candle) {
		for i := 0; i < candle.Length(); i++ {
			t.Logf("[%d] %s %f %f %f %f %f", i,
				time.Unix(candle.Timestamp[i], 0),
				candle.Open[i],
				candle.High[i],
				candle.Low[i],
				candle.Close[i],
				candle.Volume[i],
			)
		}
	}

	tests := []struct {
		Settle   string
		Contract string
		From     int64
		To       int64
		Limit    int
		Period   time.Duration
	}{
		{"usdt", "BTC_USDT", 0, 0, 10, time.Second * 10},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Minute},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Minute * 5},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Minute * 15},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Minute * 30},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Hour},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Hour * 4},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Hour * 8},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Hour * 24},
		{"usdt", "ETH_USDT", 0, 0, 10, time.Hour * 24 * 7},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.Contract, getInterval(tt.Period)), func(t *testing.T) {
			candle, err := futures.Candle(ctx, tt.Settle, tt.Contract, tt.From, tt.To, tt.Limit, tt.Period)
			require.NoError(t, err)
			print(t, candle)
		})
	}
}

func TestFutures_CommonLiquidation(t *testing.T) {
	ctx := context.TODO()
	liquidations, err := futures.CommonLiquidation(ctx, "usdt", "BTC_USDT", 0, 0, 10)
	require.NoError(t, err)
	for i, l := range liquidations {
		b, err1 := json.MarshalIndent(l, "", "  ")
		require.NoError(t, err1)
		t.Logf("[%d]\n%s", i, string(b))
	}
}
