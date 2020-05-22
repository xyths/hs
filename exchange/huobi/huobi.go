package huobi

import (
	"context"
	"errors"
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/client"
	"github.com/huobirdcenter/huobi_golang/pkg/client/accountwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/marketwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/orderwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/client/websocketclientbase"
	"github.com/huobirdcenter/huobi_golang/pkg/getrequest"
	"github.com/huobirdcenter/huobi_golang/pkg/postrequest"
	"github.com/huobirdcenter/huobi_golang/pkg/response/account"
	"github.com/huobirdcenter/huobi_golang/pkg/response/auth"
	"github.com/shopspring/decimal"
	"github.com/xyths/hs/convert"
	"log"
)

const (
	Name        string = "huobi"
	DefaultHost        = "api.huobi.me"
)

type Client struct {
	Label     string
	AccessKey string
	SecretKey string
	Host      string

	SpotAccountId int64
}

func New(label, accessKey, secretKey, host string) *Client {
	c := &Client{
		Label:     label,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
	if host != "" {
		c.Host = host
	} else {
		c.Host = DefaultHost
	}
	accountId, err := c.GetSpotAccountId()
	if err == nil {
		c.SpotAccountId = accountId
	} else {
		log.Fatalf("error when get spot account id: %s", err)
	}
	return c
}

func (c *Client) GetTimestamp() (int, error) {
	hb := new(client.CommonClient).Init(c.Host)
	return hb.GetTimestamp()
}

func (c *Client) GetAccountInfo() ([]account.AccountInfo, error) {
	hb := new(client.AccountClient).Init(c.AccessKey, c.SecretKey, c.Host)
	return hb.GetAccountInfo()
}

func (c *Client) GetSpotAccountId() (int64, error) {
	accounts, err := c.GetAccountInfo()
	if err != nil {
		return 0, err
	}
	for _, a := range accounts {
		if a.Type == "spot" {
			return a.Id, nil
		}
	}
	return 0, nil
}

func (c *Client) PlaceOrder(orderType, symbol, clientOrderId string, price, amount decimal.Decimal) (uint64, error) {
	hb := new(client.OrderClient).Init(c.AccessKey, c.SecretKey, c.Host)
	request := postrequest.PlaceOrderRequest{
		AccountId:     fmt.Sprintf("%d", c.SpotAccountId),
		Type:          orderType,
		Source:        "spot-api",
		Symbol:        symbol,
		Price:         price.String(),
		Amount:        amount.String(),
		ClientOrderId: clientOrderId,
	}
	resp, err := hb.PlaceOrder(&request)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	switch resp.Status {
	case "ok":
		log.Printf("Place order successfully, order id: %s\n", resp.Data)
		return convert.StrToUint64(resp.Data), nil
	case "error":
		log.Printf("Place order error: %s\n", resp.ErrorMessage)
		if resp.ErrorCode == "account-frozen-balance-insufficient-error" {
			return 0, nil
		}
		return 0, errors.New(resp.ErrorMessage)
	}
	return 0, errors.New("unknown status")
}

func (c *Client) SubscribeLast24hCandlestick(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) error {
	hb := new(marketwebsocketclient.Last24hCandlestickWebSocketClient).Init(c.Host)
	hb.SetHandler(
		// Connected handler
		func() {
			hb.Subscribe(symbol, clientId)
		},
		responseHandler)

	hb.Connect(true)

	select {
	case <-ctx.Done():
		hb.UnSubscribe(symbol, clientId)
		log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
	}
	return nil
}

func (c *Client) SubscribeCandlestick(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) error {
	hb := new(marketwebsocketclient.CandlestickWebSocketClient).Init(c.Host)
	hb.SetHandler(
		// Connected handler
		func() {
			hb.Subscribe(symbol, getrequest.MIN1, clientId)
		},
		responseHandler)

	hb.Connect(true)

	select {
	case <-ctx.Done():
		hb.UnSubscribe(symbol, getrequest.MIN1, clientId)
		log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
	}
	return nil
}

func (c *Client) SubscribeOrder(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) error {
	hb := new(orderwebsocketclient.SubscribeOrderWebSocketV2Client).Init(c.AccessKey, c.SecretKey, c.Host)

	hb.SetHandler(
		// Connected handler
		func(resp *auth.WebSocketV2AuthenticationResponse) {
			if resp.IsSuccess() {
				// Subscribe if authentication passed
				hb.Subscribe(symbol, clientId)
			} else {
				applogger.Error("Authentication error, code: %d, message:%s", resp.Code, resp.Message)
			}
		},
		responseHandler)

	hb.Connect(true)

	select {
	case <-ctx.Done():
		hb.UnSubscribe(symbol, clientId)
		log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
	}
	return nil
}


func (c *Client) SubscribeAccountUpdate(ctx context.Context, symbol, clientId string,
	responseHandler websocketclientbase.ResponseHandler) error {
	hb := new(accountwebsocketclient.SubscribeAccountWebSocketV2Client).Init(c.AccessKey, c.SecretKey, c.Host)

	hb.SetHandler(
		// Connected handler
		func(resp *auth.WebSocketV2AuthenticationResponse) {
			if resp.IsSuccess() {
				// Subscribe if authentication passed
				hb.Subscribe("1", clientId)
			} else {
				applogger.Error("Authentication error, code: %d, message:%s", resp.Code, resp.Message)
			}
		},
		responseHandler)

	hb.Connect(true)

	select {
	case <-ctx.Done():
		hb.UnSubscribe("1", clientId)
		log.Printf("UnSubscribed, symbol = %s, clientId = %s", symbol, clientId)
	}
	return nil
}