package gateio

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSpotV4_Currencies(t *testing.T) {
	ctx := context.Background()
	currencies, err := g4.Currencies(ctx)
	require.NoError(t, err)
	for i, c := range currencies {
		t.Logf("[%d] %v", i, c)
	}
}

func TestSpotV4_CandleBySize(t *testing.T) {
	ctx := context.Background()

	t.Run("range and size", func(t *testing.T) {
		tests := []int{1, 999, 1000, 1001, 1999, 2000, 2001, 5000}

		for i, size := range tests {
			candle, err := g4.CandleBySize(ctx, "btc_usdt", time.Hour, size)
			require.NoError(t, err)
			if candle.Length() != size {
				t.Errorf("candle size expect %d, acutal %d", size, candle.Length())
			}
			for j := 0; j < candle.Length(); j++ {
				t.Logf("[%d] [%d] %d %f %f %f %f %f", i, j,
					candle.Timestamp[j], candle.Open[j], candle.High[j], candle.Low[j], candle.Close[j], candle.Volume[j])
			}
		}
	})

	// Gate日线是从北京时间8点开始的
	t.Run("timestamp of D", func(t *testing.T) {
		candle, err := g4.CandleBySize(ctx, "btc_usdt", time.Hour*24, 10)
		require.NoError(t, err)
		for j := 0; j < candle.Length(); j++ {
			t.Logf("[%d] %d %f %f %f %f %f", j,
				candle.Timestamp[j], candle.Open[j], candle.High[j], candle.Low[j], candle.Close[j], candle.Volume[j])
		}
	})

	// weekly candle
	t.Run("timestamp of W", func(t *testing.T) {
		candle, err := g4.CandleBySize(ctx, "btc_usdt", time.Hour*24*7, 10)
		require.NoError(t, err)
		for j := 0; j < candle.Length(); j++ {
			t.Logf("[%d] %d %f %f %f %f %f", j,
				candle.Timestamp[j], candle.Open[j], candle.High[j], candle.Low[j], candle.Close[j], candle.Volume[j])
		}
	})
}

func TestSpotV4_AllSymbols(t *testing.T) {
	ctx := context.Background()
	symbols, err := g4.AllSymbols(ctx)
	require.NoError(t, err)
	for i, s := range symbols {
		b, _ := json.Marshal(s)
		t.Logf("[%d] %s", i, string(b))
	}
}
