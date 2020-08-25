package broadcast

import "fmt"

type Broadcaster interface {
	Broadcast(labels []string, symbol, time, direction, price, amount, total, profit string)
	SendText(message string) error
}

const (
	Buy  = "buy"
	Sell = "sell"

	NameDingTalk = "dingtalk"
)

type Config struct {
	Name    string
	BaseUrl string `json:"baseUrl"`
	Secret  string
}

func New(config Config) Broadcaster {
	switch config.Name {
	case NameDingTalk:
		return NewDingTalk(config)
	default:
		panic(fmt.Sprintf("robot %s not supported", config.Name))
	}
}
