package huobi

import (
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/client/marketwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
	"github.com/xyths/hs/exchange"
)

func (c *Client) SubscribeTrade(symbol, clientId string, responseHandler exchange.TradeHandler) {
	hb := new(marketwebsocketclient.TradeWebSocketClient).Init(c.Host)
	hb.SetHandler(
		// Connected handler
		func() {
			hb.Subscribe(symbol, clientId)
		},
		tradeHandler(responseHandler),
	)

	hb.Connect(true)
}

func (c *Client) UnsubscribeTrade(symbol, clientId string) {
	hb := new(marketwebsocketclient.TradeWebSocketClient).Init(c.Host)
	hb.UnSubscribe(symbol, clientId)
}

func tradeHandler(responseHandler exchange.TradeHandler) websocketclientbase.ResponseHandler {
	return func(response interface{}) {
		depthResponse, ok := response.(market.SubscribeTradeResponse)
		if ok {
			if &depthResponse != nil {
				if depthResponse.Tick != nil && depthResponse.Tick.Data != nil {
					applogger.Info("WebSocket received trade update: count=%d", len(depthResponse.Tick.Data))
					var details []exchange.TradeDetail
					l := len(depthResponse.Tick.Data)
					for i := l-1; i >= 0; i-- { // 火币的交易明细是时间倒序的，新数据在前
						t := depthResponse.Tick.Data[i]
						details = append(details, exchange.TradeDetail{
							Id:        t.TradeId,
							Price:     t.Price,
							Amount:    t.Amount,
							Timestamp: t.Timestamp,
							Direction: t.Direction,
						})
					}
					responseHandler(details)
				}

				if depthResponse.Data != nil {
					applogger.Info("WebSocket received trade data: count=%d", len(depthResponse.Data))
					//for _, t := range depthResponse.Data {
					//	applogger.Info("Trade data, id: %d, price: %v, amount: %v", t.TradeId, t.Price, t.Amount)
					//}
				}
			}
		} else {
			applogger.Warn("Unknown response: %v", response)
		}
	}
}
