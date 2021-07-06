package gateio

import (
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/xyths/hs"
	"github.com/xyths/hs/exchange"
	"os"
	"strconv"
	"testing"
	"time"
)

var g *GateIO
var g4 *SpotV4
var futures *Futures

func TestMain(m *testing.M) {
	l, err := hs.NewZapLogger(hs.LogConf{
		Level:   "debug",
		Outputs: []string{"stdout"},
		Errors:  []string{"stderr"},
	},
	)
	if err != nil {
		return
	}
	defer l.Sync()

	apiKey := os.Getenv("apiKey")
	secretKey := os.Getenv("secretKey")
	host := os.Getenv("host")

	v4apiKey := os.Getenv("v4api")
	v4secretKey := os.Getenv("v4secret")
	g = New(apiKey, secretKey, host, l.Sugar())

	g4 = NewSpotV4(v4apiKey, v4secretKey, host, l.Sugar())

	futures = NewFutures(v4apiKey, v4secretKey, host, l.Sugar())
	os.Exit(m.Run())
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_GetPairs(t *testing.T) {
	if pairs, err := g.GetPairs(); err != nil {
		t.Logf("error when GetPairs: %s", err)
	} else {
		t.Logf("GetPairs: %s", pairs)
	}
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_MarketInfo(t *testing.T) {
	//apiKey := os.Getenv("apiKey")
	//secretKey := os.Getenv("secretKey")
	//host := os.Getenv("host")
	//t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	//g := New(apiKey, secretKey, host)

	pairs, err := g.MarketInfo()

	require.NoError(t, err)
	b, err := json.MarshalIndent(pairs, "", "\t")
	require.NoError(t, err)
	t.Logf("GetPairs: %s", string(b))
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_AllSymbols(t *testing.T) {
	//apiKey := os.Getenv("apiKey")
	//secretKey := os.Getenv("secretKey")
	//host := os.Getenv("host")
	//t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	//g := New(apiKey, secretKey, host)

	symbols, err := g.AllSymbols(context.TODO())

	require.NoError(t, err)
	b, err := json.MarshalIndent(symbols, "", "\t")
	require.NoError(t, err)
	t.Logf("GetPairs: %s", string(b))
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_GetSymbol(t *testing.T) {
	//apiKey := os.Getenv("apiKey")
	//secretKey := os.Getenv("secretKey")
	//host := os.Getenv("host")
	//t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	//g := New(apiKey, secretKey, host)

	tests := []exchange.Symbol{
		{Symbol: "sero_usdt", Disabled: false, BaseCurrency: "SERO", QuoteCurrency: "USDT", PricePrecision: 5, AmountPrecision: 3, LimitOrderMinAmount: decimal.NewFromFloat(0.0001), MinTotal: decimal.NewFromFloat(1)},
		{Symbol: "btc_usdt", BaseCurrency: "BTC", QuoteCurrency: "USDT", PricePrecision: 2, AmountPrecision: 4, LimitOrderMinAmount: decimal.NewFromFloat(0.0001), MinTotal: decimal.NewFromFloat(1)},
		{Symbol: "btc3l_usdt", BaseCurrency: "BTC3L", QuoteCurrency: "USDT", PricePrecision: 4, AmountPrecision: 3, LimitOrderMinAmount: decimal.NewFromFloat(0.0001), MinTotal: decimal.NewFromFloat(1)},
		{Symbol: "btc3s_usdt", BaseCurrency: "BTC3S", QuoteCurrency: "USDT", PricePrecision: 4, AmountPrecision: 3, LimitOrderMinAmount: decimal.NewFromFloat(0.0001), MinTotal: decimal.NewFromFloat(1)},
		{Symbol: "ampl_usdt", BaseCurrency: "AMPL", QuoteCurrency: "USDT", PricePrecision: 3, AmountPrecision: 4, LimitOrderMinAmount: decimal.NewFromFloat(0.0001), MinTotal: decimal.NewFromFloat(1)},
	}
	for _, tt := range tests {
		t.Run(tt.Symbol, func(t *testing.T) {
			actual, err := g.GetSymbol(context.TODO(), tt.Symbol)
			require.NoError(t, err)
			if tt.BaseCurrency != actual.BaseCurrency {
				t.Errorf("base currency expect %s, actual %s", tt.BaseCurrency, actual.BaseCurrency)
			}
			if tt.QuoteCurrency != actual.QuoteCurrency {
				t.Errorf("quote currency expect %s, actual %s", tt.QuoteCurrency, actual.QuoteCurrency)
			}
			if tt.PricePrecision != actual.PricePrecision {
				t.Errorf("price precision expect %d, actual %d", tt.PricePrecision, actual.PricePrecision)
			}
			if tt.AmountPrecision != actual.AmountPrecision {
				t.Errorf("amount precision expect %d, actual %d", tt.AmountPrecision, actual.AmountPrecision)
			}
			if !tt.LimitOrderMinAmount.Equal(actual.LimitOrderMinAmount) {
				t.Errorf("min amount expect %s, actual %s", tt.LimitOrderMinAmount, actual.LimitOrderMinAmount)
			}
			if !tt.MinTotal.Equal(actual.MinTotal) {
				t.Errorf("min total expect %s, actual %s", tt.MinTotal, actual.MinTotal)
			}
		})
	}
}

// apiKey=xxx secretKey=yyy symbol=aaa order=1111 go test -test.v -test.run TestGateIO_GetOrderString ./gateio
func TestGateIO_GetOrderString(t *testing.T) {
	//apiKey := os.Getenv("apiKey")
	//secretKey := os.Getenv("secretKey")
	//host := os.Getenv("host")
	symbol := os.Getenv("symbol")
	orderId_ := os.Getenv("order")
	id, err := strconv.ParseUint(orderId_, 10, 64)
	require.NoError(t, err)
	//t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	//g := New(apiKey, secretKey, host)

	order, err := g.GetOrderString(id, symbol)
	require.NoError(t, err)
	require.NoError(t, err)
	t.Logf("order is %s", order)
}

// apiKey=xxx secretKey=yyy symbol=aaa order=1111 go test -test.v -test.run TestGateIO_GetOrder ./gateio
func TestGateIO_GetOrder(t *testing.T) {
	//apiKey := os.Getenv("apiKey")
	//secretKey := os.Getenv("secretKey")
	//host := os.Getenv("host")
	symbol := os.Getenv("symbol")
	orderId_ := os.Getenv("order")
	id, err := strconv.ParseUint(orderId_, 10, 64)
	require.NoError(t, err)
	//t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	//g := New(apiKey, secretKey, host)

	order, err := g.GetOrder(id, symbol)
	require.NoError(t, err)
	b, err := json.MarshalIndent(order, "", "\t")
	require.NoError(t, err)
	t.Logf("order is %s", string(b))
}

func TestGateIO_WsPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	localTime := time.Now().Unix()
	pong, err := g.ReqPing(ctx, localTime)
	require.NoError(t, err)
	expect := "pong"
	if pong != expect {
		t.Logf("expect: %s, got: %s", expect, pong)
	}
}

func TestGateIO_WsTime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	localTime := time.Now().Unix()
	serverTime, err := g.ReqTime(ctx, localTime)
	require.NoError(t, err)
	t.Logf("local: %d, server: %d", localTime, serverTime)
}

func TestGateIO_WsTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	id := time.Now().Unix()
	symbol := "BTC_USDT"
	ticker, err := g.ReqTicker(ctx, id, symbol, time.Hour)
	require.NoError(t, err)
	t.Logf("id: %d, symbol: %s, ticker: %v", id, symbol, ticker)
}

func TestGateIO_SubTicker(t *testing.T) {
	id := time.Now().Unix()
	symbol := "BTC_USDT"
	g.SubTicker(id, symbol,
		func(response interface{}) {
			t.Logf("ticker response: %v", response)
		})
	time.Sleep(1 * time.Minute)
	g.UnsubTicker(id, symbol)
}

func TestGateIO_ReqCandlestick(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	symbol := "BTC_USDT"
	end := time.Now()
	start := end.Add(-4 * time.Hour)
	candle, err := g.ReqCandlestick(ctx, symbol, "id", time.Hour, start, end)
	require.NoError(t, err)
	for i := 0; i < candle.Length(); i++ {
		t.Logf("%d %f %f %f %f %f", candle.Timestamp[i], candle.Open[i], candle.High[i], candle.Low[i], candle.Close[i], candle.Volume[i])
	}
}

func TestGateIO_SubCandlestick(t *testing.T) {
	symbol := "BTC_USDT"
	g.SubCandlestick(symbol, "id", time.Minute,
		func(response interface{}) {
			ticker, ok := response.(hs.Ticker)
			if !ok {
				t.Error("bad response format")
				return
			}
			t.Logf("%d %f %f %f %f %f", ticker.Timestamp, ticker.Open, ticker.High, ticker.Low, ticker.Close, ticker.Volume)
		})

	time.Sleep(2 * time.Minute)
	t.Log("unsubscribe")
	g.UnsubCandlestick(symbol, "id")
}

func TestPrivateWebsocketClient_ReqOrder(t *testing.T) {
	symbol := os.Getenv("symbol")
	if symbol == "" {
		symbol = "BTC_USDT"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	orders, err := g.ReqOrder(ctx, symbol, "id")
	require.NoError(t, err)
	for o := range orders {
		t.Logf("%v", o)
	}
}
