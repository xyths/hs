package base

import "github.com/huobirdcenter/huobi_golang/pkg/model/auth"

// It will be invoked after websocket connected
type ConnectedHandler func()

// It will be invoked after valid message received
type MessageHandler func(messageType int, payload []byte)

// It will be invoked after response is parsed
type ResponseHandler func(response interface{})

// It will be invoked after websocket v2 authentication response received
type AuthenticationV2ResponseHandler func(resp *auth.WebSocketV2AuthenticationResponse)
