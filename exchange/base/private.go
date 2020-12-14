package base

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	websocketV2Path = "/ws/v2"
)

// The base class that responsible to get data from websocket authentication v2
type PrivateWebSocketBase struct {
	host                string
	path                string
	TimerIntervalSecond int
	ReconnectWaitSecond int
	Logger              *zap.SugaredLogger
	verbose             bool

	conn *websocket.Conn

	authenticationResponseHandler ConnectedHandler
	messageHandler                MessageHandler

	stopReadChannel   chan int
	stopTickerChannel chan int
	ticker            *time.Ticker
	lastReceivedTime  time.Time
	sendMutex         *sync.Mutex

	auth Authentication
}

// Initializer
func (b *PrivateWebSocketBase) Init(host, path string, auth Authentication,
	logger *zap.SugaredLogger, intervalSecond, reconnectSecond int, verbose bool) *PrivateWebSocketBase {
	b.host = host
	b.path = path
	b.Logger = logger
	b.TimerIntervalSecond = intervalSecond
	b.ReconnectWaitSecond = reconnectSecond
	b.verbose = verbose
	b.stopReadChannel = make(chan int, 1)
	b.stopTickerChannel = make(chan int, 1)
	b.auth = auth
	b.sendMutex = &sync.Mutex{}
	return b
}

// Set callback handler
func (b *PrivateWebSocketBase) SetHandler(connHandler ConnectedHandler, msgHandler MessageHandler) {
	b.authenticationResponseHandler = connHandler
	b.messageHandler = msgHandler
}

// Connect to websocket server
// if autoConnect is true, then the connection can be re-connect if no data received after the pre-defined timeout
func (b *PrivateWebSocketBase) Connect(autoConnect bool) {
	b.connectWebSocket()

	if autoConnect {
		b.startTicker()
	}
}

// Send data to websocket server
func (b *PrivateWebSocketBase) Send(data string) {
	if b.conn == nil {
		applogger.Error("WebSocket sent error: no connection available")
		return
	}

	b.sendMutex.Lock()
	err := b.conn.WriteMessage(websocket.TextMessage, []byte(data))
	b.sendMutex.Unlock()

	if err != nil {
		applogger.Error("WebSocket sent error: data=%s, error=%s", data, err)
	}
}

// Close the connection to server
func (b *PrivateWebSocketBase) Close() {
	b.stopTicker()
	b.disconnectWebSocket()
}

// connect to server
func (b *PrivateWebSocketBase) connectWebSocket() {
	var err error
	url := fmt.Sprintf("wss://%s%s", b.host, websocketV2Path)
	applogger.Debug("WebSocket connecting...")
	b.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		applogger.Error("WebSocket connected error: %s", err)
		return
	}
	applogger.Info("WebSocket connected")

	auth, err := b.auth.Build()
	if err != nil {
		applogger.Error("Signature generated error: %s", err)
		return
	}

	b.Send(auth)

	b.startReadLoop()
}

// disconnect with server
func (b *PrivateWebSocketBase) disconnectWebSocket() {
	if b.conn == nil {
		return
	}

	// start a new goroutine to send a signal
	go b.stopReadLoop()

	applogger.Debug("WebSocket disconnecting...")
	err := b.conn.Close()
	if err != nil {
		applogger.Error("WebSocket disconnect error: %s", err)
		return
	}

	applogger.Info("WebSocket disconnected")
}

// initialize a ticker and start a goroutine tickerLoop()
func (b *PrivateWebSocketBase) startTicker() {
	b.ticker = time.NewTicker(time.Duration(b.TimerIntervalSecond) * time.Second)
	b.lastReceivedTime = time.Now()

	go b.tickerLoop()
}

// stop ticker and stop the goroutine
func (b *PrivateWebSocketBase) stopTicker() {
	b.ticker.Stop()
	b.stopTickerChannel <- 1
}

// defines a for loop that will run based on ticker's frequency
// It checks the last data that received from server, if it is longer than the threshold,
// it will force disconnect server and connect again.
func (b *PrivateWebSocketBase) tickerLoop() {
	applogger.Debug("tickerLoop started")
	for {
		select {
		// start a goroutine readLoop()
		case <-b.stopTickerChannel:
			applogger.Debug("tickerLoop stopped")
			return

		// Receive tick from tickChannel
		case <-b.ticker.C:
			elapsedSecond := time.Now().Sub(b.lastReceivedTime).Seconds()
			applogger.Debug("WebSocket received data %f sec ago", elapsedSecond)

			if elapsedSecond > float64(b.ReconnectWaitSecond) {
				applogger.Info("WebSocket reconnect...")
				b.disconnectWebSocket()
				b.connectWebSocket()
			}
		}
	}
}

// start a goroutine readLoop()
func (b *PrivateWebSocketBase) startReadLoop() {
	go b.readLoop()
}

// stop the goroutine readLoop()
func (b *PrivateWebSocketBase) stopReadLoop() {
	b.stopReadChannel <- 1
}

// defines a for loop to read data from server
// it will stop once it receives the signal from stopReadChannel
func (b *PrivateWebSocketBase) readLoop() {
	applogger.Debug("readLoop started")
	for {
		select {
		// Receive data from stopChannel
		case <-b.stopReadChannel:
			applogger.Debug("readLoop stopped")
			return

		default:
			if b.conn == nil {
				applogger.Error("Read error: no connection available")
				time.Sleep(time.Duration(b.TimerIntervalSecond) * time.Second)
				continue
			}

			msgType, buf, err := b.conn.ReadMessage()
			if err != nil {
				applogger.Error("Read error: %s", err)
				time.Sleep(time.Duration(b.TimerIntervalSecond) * time.Second)
				continue
			}

			b.lastReceivedTime = time.Now()
			b.messageHandler(msgType, buf)
		}
	}
}
