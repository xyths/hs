package gateio

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/xyths/hs/exchange"
	"os"
	"testing"
)

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_GetPairs(t *testing.T) {
	apiKey := os.Getenv("apiKey")
	secretKey := os.Getenv("secretKey")
	host := os.Getenv("host")
	t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	gateio := New(apiKey, secretKey, host)

	if pairs, err := gateio.GetPairs(); err != nil {
		t.Logf("error when GetPairs: %s", err)
	} else {
		t.Logf("GetPairs: %s", pairs)
	}
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_MarketInfo(t *testing.T) {
	apiKey := os.Getenv("apiKey")
	secretKey := os.Getenv("secretKey")
	host := os.Getenv("host")
	t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	g := New(apiKey, secretKey, host)

	pairs, err := g.MarketInfo()

	require.NoError(t, err)
	b, err := json.MarshalIndent(pairs, "", "\t")
	require.NoError(t, err)
	t.Logf("GetPairs: %s", string(b))
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_AllSymbols(t *testing.T) {
	apiKey := os.Getenv("apiKey")
	secretKey := os.Getenv("secretKey")
	host := os.Getenv("host")
	t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	g := New(apiKey, secretKey, host)

	symbols, err := g.AllSymbols()

	require.NoError(t, err)
	b, err := json.MarshalIndent(symbols, "", "\t")
	require.NoError(t, err)
	t.Logf("GetPairs: %s", string(b))
}

// apiKey=xxx secretKey=yyy go test -v -run TestGetPairs ./gateio
func TestGateIO_GetSymbol(t *testing.T) {
	apiKey := os.Getenv("apiKey")
	secretKey := os.Getenv("secretKey")
	host := os.Getenv("host")
	t.Logf("apiKey: %s, secretKey: %s", apiKey, secretKey)
	g := New(apiKey, secretKey, host)

	tests := []exchange.Symbol{
		{"sero_usdt", "SERO", "USDT", 5, 3, decimal.NewFromFloat(0.0001), decimal.NewFromFloat(1)},
		{"btc_usdt", "BTC", "USDT", 2, 4, decimal.NewFromFloat(0.0001), decimal.NewFromFloat(1)},
		{"btc3l_usdt", "BTC3L", "USDT", 4, 3, decimal.NewFromFloat(0.0001), decimal.NewFromFloat(1)},
		{"btc3s_usdt", "BTC3S", "USDT", 4, 3, decimal.NewFromFloat(0.0001), decimal.NewFromFloat(1)},
		{"ampl_usdt", "AMPL", "USDT", 3, 4, decimal.NewFromFloat(0.0001), decimal.NewFromFloat(1)},
	}
	for _, tt := range tests {
		t.Run(tt.Symbol, func(t *testing.T) {
			actual, err := g.GetSymbol(tt.Symbol)
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
			if !tt.MinAmount.Equal(actual.MinAmount) {
				t.Errorf("min amount expect %s, actual %s", tt.MinAmount, actual.MinAmount)
			}
			if !tt.MinTotal.Equal(actual.MinTotal) {
				t.Errorf("min total expect %s, actual %s", tt.MinTotal, actual.MinTotal)
			}
		})
	}
}
