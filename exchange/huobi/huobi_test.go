package huobi

import (
	"context"
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/response/account"
	"github.com/huobirdcenter/huobi_golang/pkg/response/market"
	"github.com/huobirdcenter/huobi_golang/pkg/response/order"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestClient_GetTimestamp(t *testing.T) {
	c := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	timestamp, err := c.GetTimestamp()
	require.NoError(t, err)
	t.Logf("timestamp is: %d", timestamp)
}

func TestClient_GetAccountInfo(t *testing.T) {
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	accounts, err := client.GetAccountInfo();
	require.NoError(t, err)

	for _, a := range accounts {
		t.Logf("account is: %#v", a)
	}
}

func TestClient_GetSpotAccountId(t *testing.T) {
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	t.Logf("spot account id is: %d", client.SpotAccountId)
}

// ACCESS_KEY=xxxx SECRET_KEY=xxxx go test ./ -v -run=TestClient_PlaceOrder
func TestClient_PlaceOrder(t *testing.T) {
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	price := decimal.NewFromFloat(8000.1)
	amount := decimal.NewFromFloat(0.001)
	clientOrderId := fmt.Sprintf("%d", time.Now().Unix())
	orderId, err := client.PlaceOrder(OrderTypeBuyLimit, BTC_USDT, clientOrderId, price, amount)
	require.NoError(t, err)
	t.Logf("place buy order, id = %d", orderId)
}

func TestClient_SubscribeLast24hCandlestick(t *testing.T) {
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))

	// Set the callback handlers
	err := client.SubscribeLast24hCandlestick(context.Background(), BTC_USDT, "1608",
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

func TestClient_SubscribeCandlestick(t *testing.T) {
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))

	// Set the callback handlers
	client.SubscribeCandlestick(context.Background(), BTC_USDT, "1101",
		func(resp interface{}) {
			candlestickResponse, ok := resp.(market.SubscribeCandlestickResponse)
			if ok {
				if &candlestickResponse != nil {
					if candlestickResponse.Tick != nil {
						t := candlestickResponse.Tick
						applogger.Info("Candlestick update, id: %d, count: %v, volume: %v, OHLC[%v-%v-%v-%v]",
							t.Id, t.Count, t.Vol, t.Open, t.High, t.Low, t.Close)
					}

					if candlestickResponse.Data != nil {
						for i, t := range (candlestickResponse.Data) {
							applogger.Info("Candlestick data[%d], id: %d, count: %v, volume: %v, OHLC[%v-%v-%v-%v]",
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
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))

	// Set the callback handlers
	//fmt.Sprintln(time.Now().Unix())
	client.SubscribeCandlestickWithReq(context.Background(), BTC_USDT, "1111",
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
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))

	// Set the callback handlers
	client.SubscribeOrder(context.Background(), BTC_USDT, "a123",
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
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))

	// Set the callback handlers
	client.SubscribeOrder(context.Background(), BTC_USDT, fmt.Sprintln(time.Now().Unix()),
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
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	price, err := client.GetPrice(BTC_USDT)
	require.NoError(t, err)
	t.Logf("lastest BTC price is: %s", price)
}

func TestClient_GetSpotBalance(t *testing.T) {
	client := New("test", os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), os.Getenv("HUOBI_HOST"))
	balance, err := client.GetSpotBalance()
	require.NoError(t, err)
	for k, v := range balance {
		t.Logf("%s:%s", k, v)
	}
}
