package base

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
	"time"
)

type WebsocketBase struct {
	host                string
	path                string
	TimerIntervalSecond int
	ReconnectWaitSecond int
	Logger              *zap.SugaredLogger
	verbose             bool

	conn              *websocket.Conn
	connectedHandler  ConnectedHandler
	messageHandler    MessageHandler
	stopReadChannel   chan int
	stopTickerChannel chan int
	ticker            *time.Ticker
	lastReceivedTime  time.Time
	sendMutex         *sync.Mutex
}

// Initializer
func (b *WebsocketBase) Init(host, path string, logger *zap.SugaredLogger, intervalSecond, reconnectSecond int, verbose bool) *WebsocketBase {
	b.host = host
	b.path = path
	b.Logger = logger
	b.TimerIntervalSecond = intervalSecond
	b.ReconnectWaitSecond = reconnectSecond
	b.verbose = verbose
	b.stopReadChannel = make(chan int, 1)
	b.stopTickerChannel = make(chan int, 1)
	b.sendMutex = &sync.Mutex{}

	return b
}

// Set callback handler
func (b *WebsocketBase) SetHandler(connHandler ConnectedHandler, msgHandler MessageHandler) {
	b.connectedHandler = connHandler
	b.messageHandler = msgHandler
}

// Connect to websocket server
// if autoConnect is true, then the connection can be re-connect if no data received after the pre-defined timeout
func (b *WebsocketBase) Connect(autoConnect bool) {
	b.connectWebSocket()

	if autoConnect {
		b.startTicker()
	}
}

// Send data to websocket server
func (b *WebsocketBase) Send(data string) {
	if b.conn == nil {
		if b.verbose {
			b.Logger.Error("WebSocket sent error: no connection available")
		}
		return
	}

	b.sendMutex.Lock()
	err := b.conn.WriteMessage(websocket.TextMessage, []byte(data))
	b.sendMutex.Unlock()

	if err != nil {
		if b.verbose {
			b.Logger.Error("WebSocket sent error: data=%s, error=%s", data, err)
		}
	}
}

// Close the connection to server
func (b *WebsocketBase) Close() {
	b.stopTicker()
	b.disconnectWebSocket()
}

// connect to server
func (b *WebsocketBase) connectWebSocket() {
	var err error
	url := fmt.Sprintf("wss://%s%s", b.host, b.path)
	if b.verbose {
		b.Logger.Debug("WebSocket connecting...")
	}
	b.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		if b.verbose {
			b.Logger.Errorf("WebSocket connected error: %s", err)
		}
		return
	}
	if b.verbose {
		b.Logger.Info("WebSocket connected")
	}

	b.startReadLoop()

	if b.connectedHandler != nil {
		b.connectedHandler()
	}
}

// disconnect with server
func (b *WebsocketBase) disconnectWebSocket() {
	if b.conn == nil {
		return
	}

	b.stopReadLoop()

	if b.verbose {
		b.Logger.Debug("WebSocket disconnecting...")
	}
	err := b.conn.Close()
	if err != nil {
		if b.verbose {
			b.Logger.Error("WebSocket disconnect error: %s", err)
		}
		return
	}

	if b.verbose {
		b.Logger.Info("WebSocket disconnected")
	}
}

// initialize a ticker and start a goroutine tickerLoop()
func (b *WebsocketBase) startTicker() {
	b.ticker = time.NewTicker(time.Duration(b.TimerIntervalSecond) * time.Second)
	b.lastReceivedTime = time.Now()

	go b.tickerLoop()
}

// stop ticker and stop the goroutine
func (b *WebsocketBase) stopTicker() {
	b.ticker.Stop()
	b.stopTickerChannel <- 1
}

// defines a for loop that will run based on ticker's frequency
// It checks the last data that received from server, if it is longer than the threshold,
// it will force disconnect server and connect again.
func (b *WebsocketBase) tickerLoop() {
	if b.verbose {
		b.Logger.Debug("tickerLoop started")
	}
	for {
		select {
		// Receive data from stopChannel
		case <-b.stopTickerChannel:
			if b.verbose {
				b.Logger.Debug("tickerLoop stopped")
			}
			return

		// Receive tick from tickChannel
		case <-b.ticker.C:
			elapsedSecond := time.Now().Sub(b.lastReceivedTime).Seconds()
			if b.verbose {
				b.Logger.Debugf("WebSocket received data %f sec ago", elapsedSecond)
			}

			if elapsedSecond > float64(b.ReconnectWaitSecond) {
				if b.verbose {
					b.Logger.Info("WebSocket reconnect...")
				}
				b.disconnectWebSocket()
				b.connectWebSocket()
			}
		}
	}
}

// start a goroutine readLoop()
func (b *WebsocketBase) startReadLoop() {
	go b.readLoop()
}

// stop the goroutine readLoop()
func (b *WebsocketBase) stopReadLoop() {
	b.stopReadChannel <- 1
}

// defines a for loop to read data from server
// it will stop once it receives the signal from stopReadChannel
func (b *WebsocketBase) readLoop() {
	if b.verbose {
		b.Logger.Debug("readLoop started")
	}
	for {
		select {
		// Receive data from stopChannel
		case <-b.stopReadChannel:
			if b.verbose {
				b.Logger.Debug("readLoop stopped")
			}
			return
		//case <-time.After(time.Duration(b.TimerIntervalSecond) * time.Second):
		//	if b.conn == nil {
		//		if b.verbose {
		//			b.Logger.Error("Read error: no connection available")
		//		}
		//		continue
		//	}
		//	msgType, buf, err := b.conn.ReadMessage()
		//	if err != nil {
		//		if b.verbose {
		//			b.Logger.Errorf("Read error: %s", err)
		//		}
		//		continue
		//	}
		//
		//	b.lastReceivedTime = time.Now()
		//	b.messageHandler(msgType, buf)

		default:
			if b.conn == nil {
				if b.verbose {
					b.Logger.Error("Read error: no connection available")
				}
				time.Sleep(time.Duration(b.TimerIntervalSecond) * time.Second)
				continue
			}

			msgType, buf, err := b.conn.ReadMessage()
			if err != nil {
				if b.verbose {
					b.Logger.Errorf("Read error: %s", err)
				}
				time.Sleep(time.Duration(b.TimerIntervalSecond) * time.Second)
				continue
			}

			b.lastReceivedTime = time.Now()
			b.messageHandler(msgType, buf)
		}
	}
}
