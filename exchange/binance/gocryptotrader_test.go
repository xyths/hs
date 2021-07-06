package binance

import (
	"github.com/stretchr/testify/require"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/binance"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"testing"
	"time"
)

func TestBinanceExchange(t *testing.T) {
	var b binance.Binance
	b.SetDefaults()
	t.Run("get kline", func(t *testing.T) {
		candles, err := b.GetSpotKline(&binance.KlinesRequestParams{
			Symbol:   currency.NewPairWithDelimiter("BTC", "USDT", "/"),
			Interval: kline.FiveMin.Short(),
			Limit:    24,
			//StartTime: time.Unix(1577836800, 0),
			EndTime: time.Now(),
		})
		if err != nil {
			t.Error("Binance GetSpotKline() error", err)
		}
		for i, c := range candles {
			t.Logf("[%d] %s %f %f %f %f %f", i, c.OpenTime, c.Open, c.High, c.Low, c.Close, c.Volume)
		}
	})
	t.Run("get candle extended", func(t *testing.T) {
		pair := currency.NewPairWithDelimiter("BTC", "USDT", "/")
		end := time.Now()
		interval := time.Hour
		start := end.Add(-10 * interval)
		res, err := b.GetHistoricCandlesExtended(pair, asset.Spot, start, end, kline.Interval(interval))
		require.NoError(t, err)
		for i, c := range res.Candles {
			t.Logf("[%d] %s %f %f %f %f %f", i, c.Time, c.Open, c.High, c.Low, c.Close, c.Volume)
		}
	})
}
