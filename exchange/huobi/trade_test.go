package huobi

import (
	"github.com/xyths/hs/exchange"
	"log"
	"testing"
	"time"
)

func TestClient_SubscribeTrade(t *testing.T) {
	handler := func(trades []exchange.TradeDetail) {
		for i, td := range trades {
			t.Logf("[%d] %v", i, td)
		}
	}
	symbol := "btcusdt"
	clientId := "tradetest"
	c.SubscribeTrade(symbol, clientId, handler)
	log.Println("subscribed")
	defer func() {
		c.UnsubscribeTrade(symbol, clientId)
		log.Println("unsubscribed")
	}()

	time.Sleep(time.Minute * 1)
}
