package gateio
//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/huobirdcenter/huobi_golang/logging/applogger"
//	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
//	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
//	"github.com/xyths/hs/exchange/base"
//	"go.uber.org/zap"
//)
//
//// Responsible to handle candlestick data from WebSocket
//type CandlestickWebSocketClient struct {
//	base.WebsocketBase
//	responseHandler base.ResponseHandler
//}
//
//// Initializer
//func (p *CandlestickWebSocketClient) Init(host, path string, logger *zap.SugaredLogger) *CandlestickWebSocketClient {
//	p.WebsocketBase.Init(host, path, logger, WsIntervalSecond, WsReconnectSecond, true)
//	return p
//}
//
//// Set callback handler
//func (p *CandlestickWebSocketClient) SetHandler(
//	connectedHandler websocketclientbase.ConnectedHandler,
//	responseHandler websocketclientbase.ResponseHandler) {
//	c.responseHandler = responseHandler
//
//	p.WebsocketBase.SetHandler(connectedHandler, p.handleMessage)
//}
//
//// Request the full candlestick data according to specified criteria
//func (p *CandlestickWebSocketClient) Request(symbol string, period string, from int64, to int64, clientId string) {
//	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
//	req := fmt.Sprintf("{\"req\": \"%s\", \"from\":%d, \"to\":%d, \"id\": \"%s\" }", topic, from, to, clientId)
//
//	p.Send(req)
//
//	applogger.Info("WebSocket requested, topic=%s, clientId=%s", topic, clientId)
//}
//
//// Subscribe candlestick data
//func (p *CandlestickWebSocketClient) Subscribe(symbol string, period string, clientId string) {
//	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
//	sub := fmt.Sprintf("{\"sub\": \"%s\", \"id\": \"%s\"}", topic, clientId)
//
//	p.Send(sub)
//
//	applogger.Info("WebSocket subscribed, topic=%s, clientId=%s", topic, clientId)
//}
//
//// Unsubscribe candlestick data
//func (p *CandlestickWebSocketClient) UnSubscribe(symbol string, period string, clientId string) {
//	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
//	unsub := fmt.Sprintf("{\"unsub\": \"%s\", \"id\": \"%s\" }", topic, clientId)
//
//	p.Send(unsub)
//
//	applogger.Info("WebSocket unsubscribed, topic=%s, clientId=%s", topic, clientId)
//}
//
//func (p *CandlestickWebSocketClient) handleMessage(msg string) (interface{}, error) {
//	result := market.SubscribeCandlestickResponse{}
//	err := json.Unmarshal([]byte(msg), &result)
//	return result, err
//}
