package hs

import (
	"encoding/json"
	"github.com/xyths/hs/broadcast"
	"os"
)

type MongoConf struct {
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
	URI         string `json:"uri"`
	Database    string `json:"database"`
	MaxPoolSize uint64 `json:"maxPoolSize"`
	MinPoolSize uint64 `json:"minPoolSize"`
	AppName     string `json:"appName"`
}

type MySQLConf struct {
	URI string `json:"uri"`
}

type SQLiteConf struct {
	Location string
}

type ExchangeConf struct {
	Name    string // see const below
	Label   string
	Symbols []string
	Key     string
	Secret  string
	Host    string
}

type BroadcastConf = broadcast.Config

const (
	GateIO = "gate"
	MXC    = "mxc"
	OKEx   = "okex"
	Huobi  = "huobi"
	Binance = "binance"
)

type GridStrategyConf struct {
	MaxPrice float64
	MinPrice float64
	Number   int
	Total    float64
}

type RestGridStrategyConf struct {
	MaxPrice  float64
	MinPrice  float64
	Number    int
	Total     float64
	Rebalance bool
	Interval  string // sleep interval
}

type HistoryConf struct {
	Prefix   string
	Interval string
}

type LogConf struct {
	Level   string
	Outputs []string
	Errors  []string
}

type GinConf struct {
	Listen string // "host:port"
	Log    string
}

func ParseJsonConfig(filename string, config interface{}) error {
	configFile, err := os.Open(filename)
	defer func() {
		_ = configFile.Close()
	}()
	if err != nil {
		return err
	}
	err = json.NewDecoder(configFile).Decode(config)
	if err != nil {
		return err
	}
	return nil
}
