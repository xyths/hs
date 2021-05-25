// V2 Websocket

package gateio

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xyths/hs/exchange"
	"github.com/xyths/hs/exchange/base"
	"go.uber.org/zap"
	"hash"
	"strings"
	"time"
)

const (
	// 使用特殊的消息ID区分鉴权消息和普通订阅消息
	authId = 100
)

type GateAuthentication struct {
	api    string
	secret string

	hash hash.Hash
}

func (a *GateAuthentication) Init(apiKey, secretKey string) {
	a.api = apiKey
	a.secret = secretKey
	a.hash = hmac.New(sha512.New, []byte(a.secret))
}

func (a *GateAuthentication) Build() (string, error) {
	a.hash.Reset()
	nonce := time.Now().Unix() * 1000
	a.hash.Write([]byte(fmt.Sprintf("%d", nonce)))
	signature := fmt.Sprintf("%x", a.hash.Sum(nil))

	req := WebsocketRequest{
		Id:     authId,
		Method: "server.sign",
		Params: []interface{}{
			a.api, signature, nonce,
		},
	}
	return req.String(), nil
}

type WebsocketClient struct {
	base.WebsocketBase
	//base.WebsocketBase
	responseHandler base.ResponseHandler
}

// Initializer
func (c *WebsocketClient) Init(host, path string, logger *zap.SugaredLogger) *WebsocketClient {
	c.WebsocketBase.Init(host, path, logger, 5, 60, true)
	return c
}

// Set callback handler
func (c *WebsocketClient) SetHandler(
	connectedHandler base.ConnectedHandler,
	responseHandler base.ResponseHandler) {
	c.responseHandler = responseHandler
	c.WebsocketBase.SetHandler(connectedHandler, c.handleMessage)
}

func (c *WebsocketClient) handleMessage(messageType int, payload []byte) {
	// decompress gzip data if it is binary message
	switch messageType {
	case websocket.BinaryMessage:
		c.Logger.Debugf("got binary message, len = %d", len(payload))
	case websocket.TextMessage:
		msg := string(payload)
		c.Logger.Debugf("got text message: %s", msg)
		if strings.Contains(msg, "result") {
			c.handleReqMessage(payload)
		} else if strings.Contains(msg, "method") {
			c.handleSubMessage(payload)
		}
	default:
		c.Logger.Debugf("message type: %v", messageType)
	}
}

func (c *WebsocketClient) handleReqMessage(payload []byte) {
	var r ResponseWsBase
	if err := json.Unmarshal(payload, &r); err != nil {
		c.Logger.Errorf("Unmarshal response error: %s", err)
		return
	}
	if r.Error != nil {
		c.Logger.Infof("response error: %v", r.Error)
		return
	}
	if r.Id == authId {
		c.Logger.Info("auth success")
		if c.responseHandler != nil {
			c.responseHandler(r.Result)
		}
		return
	}
	// handle subscribe success, early return
	if success, ok := r.Result.(map[string]interface{}); ok {
		if success["status"] == "success" {
			c.Logger.Infof("Subscribe successful, id %d", r.Id)
			return
		}
	}
	if c.responseHandler != nil {
		c.responseHandler(r.Result)
	}
}

func (c *WebsocketClient) handleSubMessage(payload []byte) {
	var b UpdateWsBase
	if err := json.Unmarshal(payload, &b); err != nil {
		c.Logger.Errorf("Unmarshal response error: %s", err)
		return
	}
	if b.Params == nil {
		c.Logger.Infof("response has no params")
		return
	}
	if c.responseHandler != nil {
		c.responseHandler(b.Params)
	}
}

func (c *WebsocketClient) Ping(id int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "server.ping",
		Params: make([]interface{}, 0),
	}
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) PingHandler(handler base.ResponseHandler) base.ResponseHandler {
	// do nothing
	return func(response interface{}) {
		handler(response)
	}
}

func (c *WebsocketClient) Time(id int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "server.time",
		Params: make([]interface{}, 0),
	}
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) TimeHandler(handler base.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		// string to time
		timestamp, ok := response.(float64)
		if !ok {
			c.Logger.Errorf("wrong response for time: %v", response)
			return
		}
		handler(int64(timestamp))
	}
}

func (c *WebsocketClient) ReqTicker(id int64, symbol string, period int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "ticker.query",
		Params: make([]interface{}, 0),
	}
	req.Params = append(req.Params, symbol)
	req.Params = append(req.Params, period)
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) ReqTickerHandler(handler base.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		var ticker ResponseWsTicker
		if data, err := json.Marshal(response); err != nil {
			c.Logger.Errorf("parse response error: %s", err)
			return
		} else {
			if err1 := json.Unmarshal(data, &ticker); err1 != nil {
				c.Logger.Errorf("parse response error: %s", err1)
				return
			}
		}
		handler(ticker)
	}
}

func (c *WebsocketClient) SubTicker(id int64, symbol string) {
	req := WebsocketRequest{
		Id:     id,
		Method: "ticker.subscribe",
		Params: make([]interface{}, 1),
	}
	req.Params[0] = symbol
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) UnsubTicker(id int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "ticker.unsubscribe",
		Params: make([]interface{}, 0),
	}
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) SubTickerHandler(handler exchange.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		// only one symbol
		l, ok := response.([]interface{})
		if !ok {
			return
		}
		for _, item := range l {
			switch item.(type) {
			case string:
				// ignore symbol
			case map[string]interface{}:
				var ticker ResponseWsTicker
				data, err := json.Marshal(item)
				if err != nil {
					c.Logger.Errorf("parse response error: %s", err)
					continue
				}
				if err1 := json.Unmarshal(data, &ticker); err1 != nil {
					c.Logger.Errorf("parse response error: %s", err1)
					return
				}
				handler(ticker)
			}
		}
	}
}

// start, end, interval unit is second
func (c *WebsocketClient) ReqCandle(id int64, symbol string, start, end, interval int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "kline.query",
		Params: []interface{}{
			symbol, start, end, interval,
		},
	}
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) SubCandle(id int64, symbol string, interval int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "kline.subscribe",
		Params: []interface{}{
			symbol, interval,
		},
	}
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) UnsubCandle(id int64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "kline.unsubscribe",
		Params: make([]interface{}, 0),
	}
	c.WebsocketBase.Send(req.String())
}

func (c *WebsocketClient) SubCandleHandler(handler exchange.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		r, ok := response.([]interface{})
		if !ok {
			return
		}
		tickers, err := parseTickers(r)
		if err != nil {
			c.Logger.Errorf("parse ticker error: %s", err)
			return
		}
		c.Logger.Debugf("receive %d tickers", len(tickers))
		// pass a ticker to callback
		for _, t := range tickers {
			handler(t)
		}
	}
}

func (c *WebsocketClient) ReqCandleHandler(handler base.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		r, ok := response.([]interface{})
		if !ok {
			return
		}
		candle, _ := parseCandle(r)
		handler(candle)
	}
}

func (c *WebsocketClient) Auth(api, secret string) {
	auth := GateAuthentication{}
	auth.Init(api, secret)
	a, _ := auth.Build()
	c.WebsocketBase.Send(a)
}

func (c *WebsocketClient) AuthHandler(handler base.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		handler(response)
	}
}

type PrivateWebsocketClient struct {
	base.WebsocketBase
	auth             *GateAuthentication
	connectedHandler base.ConnectedHandler
	responseHandler  base.ResponseHandler
}

// Initializer
func (c *PrivateWebsocketClient) Init(host, path string, apiKey, secretKey string, logger *zap.SugaredLogger) *PrivateWebsocketClient {
	c.auth = &GateAuthentication{}
	c.auth.Init(apiKey, secretKey)
	c.WebsocketBase.Init(host, path, logger, 5, 60, true)
	return c
}

// Set callback handler
func (c *PrivateWebsocketClient) SetHandler(
	connectedHandler base.ConnectedHandler,
	responseHandler base.ResponseHandler) {
	c.connectedHandler = connectedHandler
	c.responseHandler = responseHandler
	c.WebsocketBase.SetHandler(c.authHandler, c.handleMessage)
}

func (c *PrivateWebsocketClient) handleMessage(messageType int, payload []byte) {
	// decompress gzip data if it is binary message
	switch messageType {
	case websocket.BinaryMessage:
		c.Logger.Debugf("got binary message, len = %d", len(payload))
	case websocket.TextMessage:
		msg := string(payload)
		c.Logger.Debugf("got text message: %s", msg)
		if strings.Contains(msg, "result") {
			c.handleReqMessage(payload)
		} else if strings.Contains(msg, "method") {
			c.handleSubMessage(payload)
		}
	default:
		c.Logger.Debugf("message type: %v", messageType)
	}
}

func (c *PrivateWebsocketClient) handleReqMessage(payload []byte) {
	var r ResponseWsBase
	if err := json.Unmarshal(payload, &r); err != nil {
		c.Logger.Errorf("Unmarshal response error: %s", err)
		return
	}
	if r.Error != nil {
		c.Logger.Infof("response error: %s", r.Error)
		return
	}
	if r.Id == authId {
		c.Logger.Info("auth success")
		if c.connectedHandler != nil {
			c.connectedHandler()
		}
		return
	}
	// handle subscribe success, early return
	if success, ok := r.Result.(map[string]interface{}); ok {
		if success["status"] == "success" {
			c.Logger.Infof("Subscribe successful, id %d", r.Id)
			return
		}
	}
	if c.responseHandler != nil {
		c.responseHandler(r.Result)
	}
}

func (c *PrivateWebsocketClient) handleSubMessage(payload []byte) {
	var b UpdateWsBase
	if err := json.Unmarshal(payload, &b); err != nil {
		c.Logger.Errorf("Unmarshal response error: %s", err)
		return
	}
	if b.Params == nil {
		c.Logger.Infof("response has no params")
		return
	}
	if c.responseHandler != nil {
		c.responseHandler(b.Params)
	}
}

func (c *PrivateWebsocketClient) authHandler() {
	a, _ := c.auth.Build()
	c.WebsocketBase.Send(a)
}

func (c *PrivateWebsocketClient) ReqOrder(id int64, symbol string, offset, limit uint64) {
	req := WebsocketRequest{
		Id:     id,
		Method: "order.query",
		Params: []interface{}{
			symbol, offset, limit,
		},
	}
	c.WebsocketBase.Send(req.String())
}

func (c *PrivateWebsocketClient) SubOrder(id int64, symbols []string) {
	req := WebsocketRequest{
		Id:     id,
		Method: "order.subscribe",
		Params: make([]interface{}, 0),
	}
	for _, s := range symbols {
		req.Params = append(req.Params, s)
	}

	c.WebsocketBase.Send(req.String())
}

func (c *PrivateWebsocketClient) UnsubOrder(id int64, symbols []string) {
	req := WebsocketRequest{
		Id:     id,
		Method: "order.unsubscribe",
		Params: make([]interface{}, 0),
	}
	for _, s := range symbols {
		req.Params = append(req.Params, s)
	}
	c.WebsocketBase.Send(req.String())
}

// ReqOrderHandler cast the raw result to gate order object list
func (c *PrivateWebsocketClient) ReqOrderHandler(handler base.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		if response == nil {
			handler(nil)
			return
		}
		var r ResponseReqOrder
		data, err := json.Marshal(response)
		if err != nil {
			c.Logger.Errorf("parse response error: %s", err)
			handler(nil)
			return
		}
		if err1 := json.Unmarshal(data, &r); err1 != nil {
			c.Logger.Errorf("parse response error: %s", err1)
			handler(nil)
			return
		}
		handler(r)
		return
	}
}

// client handler got a exchange.Order as response if everything's ok
func (c *PrivateWebsocketClient) SubOrderHandler(handler exchange.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		c.Logger.Debugf("order update: %v", response)
		r, ok := response.([]interface{})
		if !ok {
			return
		}
		order, err := parseOrderUpdate(r)
		if err != nil {
			c.Logger.Errorf("parse ticker error: %s", err)
			return
		}
		c.Logger.Debugf("receive order %v update", order)
		handler(order)
	}
}

func (c *PrivateWebsocketClient) ReqBalance(id int64, assets []string) {
	req := WebsocketRequest{
		Id:     id,
		Method: "balance.query",
		Params: make([]interface{}, 0),
	}
	for _, s := range assets {
		req.Params = append(req.Params, s)
	}
	c.WebsocketBase.Send(req.String())
}

func (c *PrivateWebsocketClient) SubBalance(id int64, assets []string) {
	req := WebsocketRequest{
		Id:     id,
		Method: "balance.subscribe",
		Params: make([]interface{}, 0),
	}
	for _, s := range assets {
		req.Params = append(req.Params, s)
	}

	c.WebsocketBase.Send(req.String())
}

func (c *PrivateWebsocketClient) UnsubBalance(id int64, assets []string) {
	req := WebsocketRequest{
		Id:     id,
		Method: "balance.unsubscribe",
		Params: make([]interface{}, 0),
	}
	for _, s := range assets {
		req.Params = append(req.Params, s)
	}
	c.WebsocketBase.Send(req.String())
}

// ReqBalanceHandler cast the raw result to gate order object list
func (c *PrivateWebsocketClient) ReqBalanceHandler(handler base.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		if response == nil {
			handler(nil)
			return
		}
		var r ResponseReqOrder
		data, err := json.Marshal(response)
		if err != nil {
			c.Logger.Errorf("parse response error: %s", err)
			handler(nil)
			return
		}
		if err1 := json.Unmarshal(data, &r); err1 != nil {
			c.Logger.Errorf("parse response error: %s", err1)
			handler(nil)
			return
		}
		handler(r)
		return
	}
}

func (c *PrivateWebsocketClient) SubBalanceHandler(handler exchange.ResponseHandler) base.ResponseHandler {
	return func(response interface{}) {
		c.Logger.Debugf("order update: %v", response)
		r, ok := response.([]interface{})
		if !ok {
			return
		}
		order, err := parseOrderUpdate(r)
		if err != nil {
			c.Logger.Errorf("parse ticker error: %s", err)
			return
		}
		c.Logger.Debugf("receive order %v update", order)
		handler(order)
	}
}
