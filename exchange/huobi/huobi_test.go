package huobi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/model/account"
	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
	"github.com/huobirdcenter/huobi_golang/pkg/model/order"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
	"time"
)

var c *Client

func TestMain(m *testing.M) {
	var err error
	c, err = New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	if err != nil {
		log.Fatalf("new client error: %s", err)
	}
	os.Exit(m.Run())
}

func TestClient_GetTimestamp(t *testing.T) {
	timestamp, err := c.GetTimestamp()
	require.NoError(t, err)
	t.Logf("timestamp is: %d", timestamp)
}

func TestClient_AllSymbols(t *testing.T) {
	symbols, err := c.AllSymbols(context.Background())
	require.NoError(t, err)
	for _, s := range symbols {
		str, err := json.MarshalIndent(s, "", "\t")
		require.NoError(t, err)
		t.Logf(" %s", str)
	}
}

func TestClient_GetSymbol(t *testing.T) {
	tests := []string{"btcusdt", "ethusdt"}

	ctx := context.Background()
	for _, a := range tests {
		s, err := c.GetSymbol(ctx, a)
		require.NoError(t, err)
		str, err := json.MarshalIndent(s, "", "\t")
		require.NoError(t, err)
		t.Logf(" %s", str)
	}
}

func TestClient_GetFee(t *testing.T) {
	fee, err := c.GetFee("btcusdt")
	require.NoError(t, err)
	b, err := json.MarshalIndent(fee, "", "\t")
	require.NoError(t, err)
	t.Logf("got fee: %s", string(b))
}

func TestClient_GetAccountInfo(t *testing.T) {
	accounts, err := c.GetAccountInfo()
	require.NoError(t, err)

	for _, a := range accounts {
		t.Logf("account is: %#v", a)
	}
}

func TestClient_GetSpotAccountId(t *testing.T) {
	t.Logf("spot account id is: %d", c.SpotAccountId)
}

// ACCESS_KEY=xxxx SECRET_KEY=xxxx go test ./ -v -run=TestClient_PlaceOrder
//func TestClient_PlaceOrder(t *testing.T) {
//	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
//	price := decimal.NewFromFloat(8000.1)
//	amount := decimal.NewFromFloat(0.001)
//	clientOrderId := fmt.Sprintf("%d", time.Now().Unix())
//	orderId, err := client.PlaceOrder(OrderTypeBuyLimit, "btcusdt", clientOrderId, price, amount)
//	require.NoError(t, err)
//	t.Logf("place buy order, id = %d", orderId)
//}

func TestClient_SubscribeLast24hCandlestick(t *testing.T) {
	// Set the callback handlers
	err := c.SubscribeLast24hCandlestick(context.Background(), "btcusdt", "1608",
		func(resp interface{}) {
			candlestickResponse, ok := resp.(market.SubscribeLast24hCandlestickResponse)
			if ok {
				if &candlestickResponse != nil {
					if candlestickResponse.Tick != nil {
						t := candlestickResponse.Tick
						applogger.Info("Candlestick update, id: %d, count: %v, volume: %v [%v-%v-%v-%v]",
							t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)
					}

					if candlestickResponse.Data != nil {
						t := candlestickResponse.Data
						applogger.Info("Candlestick data, id: %d, count: %v, volume: %v [%v-%v-%v-%v]",
							t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)
					}
				}
			} else {
				applogger.Warn("Unknown response: %v", resp)
			}
		})

	require.NoError(t, err)
}

func TestClient_CandleFrom(t *testing.T) {
	period := 5 * time.Minute
	t.Run("300 candles till now", func(t *testing.T) {
		to := time.Now()
		from := to.Add(-1 * CandlestickReqMaxLength * period)
		candle, err := c.CandleFrom("btcusdt", "1101", period, from, to)
		require.NoError(t, err)
		t.Logf("candle length: %d", candle.Length())
		for i := 1; i < candle.Length(); i++ {
			if candle.Timestamp[i-1] >= candle.Timestamp[i] {
				t.Errorf("Timestamp[%d] (%d) >= [%d] (%d)", i-1, candle.Timestamp[i-1], i, candle.Timestamp[i])
			}
		}
	})
	t.Run("600 candles till now", func(t *testing.T) {
		to := time.Now()
		from := to.Add(-1 * 2 * CandlestickReqMaxLength * period)
		candle, err := c.CandleFrom("btcusdt", "1101", period, from, to)
		require.NoError(t, err)
		t.Logf("candle length: %d", candle.Length())
		for i := 1; i < candle.Length(); i++ {
			if candle.Timestamp[i-1] >= candle.Timestamp[i] {
				t.Errorf("Timestamp[%d] (%d) >= [%d] (%d)", i-1, candle.Timestamp[i-1], i, candle.Timestamp[i])
			}
		}
	})
	t.Run("1000 candles till now", func(t *testing.T) {
		to := time.Now()
		from := to.Add(-1000 * period)
		candle, err := c.CandleFrom("btcusdt", "1101", period, from, to)
		require.NoError(t, err)
		t.Logf("candle length: %d", candle.Length())
		for i := 1; i < candle.Length(); i++ {
			if candle.Timestamp[i-1] >= candle.Timestamp[i] {
				t.Errorf("Timestamp[%d] (%d) >= [%d] (%d)", i-1, candle.Timestamp[i-1], i, candle.Timestamp[i])
			}
		}
	})
}

func TestClient_SubscribeCandlestick(t *testing.T) {
	// Set the callback handlers
	c.SubscribeCandlestick("btcusdt", "1101", time.Minute,
		func(resp interface{}) {
			candlestickResponse, ok := resp.(market.SubscribeCandlestickResponse)
			if ok {
				if &candlestickResponse != nil {
					if candlestickResponse.Tick != nil {
						t := candlestickResponse.Tick
						applogger.Info("Candlestick update, id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
							t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)
					}

					if candlestickResponse.Data != nil {
						for i, t := range candlestickResponse.Data {
							applogger.Info("Candlestick data[%d], id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
								i, t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)

						}
					}
				}
			} else {
				applogger.Warn("Unknown response: %v", resp)
			}
		})
}

func TestClient_SubscribeCandlestickWithReq(t *testing.T) {
	// Set the callback handlers
	//fmt.Sprintln(time.Now().Unix())
	c.SubscribeCandlestickWithReq("btcusdt", "1111", time.Minute,
		func(resp interface{}) {
			candlestickResponse, ok := resp.(market.SubscribeCandlestickResponse)
			if ok {
				if &candlestickResponse != nil {
					if candlestickResponse.Tick != nil {
						t := candlestickResponse.Tick
						applogger.Info("Candlestick update, id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
							t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)
					}

					if candlestickResponse.Data != nil {
						for i, t := range candlestickResponse.Data {
							applogger.Info("Candlestick data[%d], id: %d, count: %v, volume: %v, OHLC[%v, %v, %v, %v]",
								i, t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)

						}
					}
				}
			} else {
				applogger.Warn("Unknown response: %v", resp)
			}
		})
}

func TestClient_SubscribeOrder(t *testing.T) {
	// Set the callback handlers
	c.SubscribeOrder("btcusdt", "a123",
		func(resp interface{}) {
			subResponse, ok := resp.(order.SubscribeOrderV2Response)
			if ok {
				applogger.Info("subResponse = %#v", subResponse)
				if subResponse.Action == "sub" {
					if subResponse.IsSuccess() {
						applogger.Info("Subscription topic %s successfully", subResponse.Ch)
					} else {
						applogger.Error("Subscription topic %s error, code: %d, message: %s", subResponse.Ch, subResponse.Code, subResponse.Message)
					}
				} else if subResponse.Action == "push" {
					if subResponse.Data != nil {
						o := subResponse.Data
						applogger.Info("Order update, event: %s, symbol: %s, type: %s, status: %s",
							o.EventType, o.Symbol, o.Type, o.OrderStatus)
					}
				}
			} else {
				applogger.Warn("Received unknown response: %v", resp)
			}
		})
}

func TestClient_SubscribeAccountUpdate(t *testing.T) {
	// Set the callback handlers
	c.SubscribeOrder("btcusdt", fmt.Sprintln(time.Now().Unix()),
		func(resp interface{}) {
			subResponse, ok := resp.(account.SubscribeAccountV2Response)
			if ok {
				applogger.Info("subResponse = %#v", subResponse)
			} else {
				applogger.Warn("Received unknown response: %v", resp)
			}
		})
}

func TestClient_GetPrice(t *testing.T) {
	price, err := c.LastPrice("btcusdt")
	require.NoError(t, err)
	t.Logf("lastest BTC price is: %s", price)
}

func TestClient_GetSpotBalance(t *testing.T) {
	balance, err := c.SpotAvailableBalance()
	require.NoError(t, err)
	for k, v := range balance {
		t.Logf("%s:%s", k, v)
	}
}
