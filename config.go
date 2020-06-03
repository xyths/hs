package hs

import (
	"encoding/json"
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

type ExchangeConf struct {
	Name    string
	Label   string
	Symbols string // btc3l_usdt|btc3s_usdt
	Key     string
	Secret  string
	Host    string
}

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
