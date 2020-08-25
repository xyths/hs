package broadcast

import (
	"fmt"
	"github.com/xyths/dingtalk-webhook/dingtalk"
	"log"
	"strings"
)

type DingTalk struct {
	conf Config
	bot  *dingtalk.Client
}

func NewDingTalk(conf Config) *DingTalk {
	return &DingTalk{
		conf: conf,
		bot:  dingtalk.New(conf.BaseUrl, conf.Secret),
	}
}

func (d *DingTalk) Broadcast(labels []string, symbol, time, direction, price, amount, total, profit string) {
	var title string
	switch direction {
	case Buy:
		title = "买入"
	case Sell:
		title = "卖出"
	}
	msg := fmt.Sprintf(`%s [%s]
[%s] [%s]
成交均价 %s, 成交量 %s, 成交额 %s, 利润 %s`, time, title, strings.Join(labels, "] ["), symbol, price, amount, total, profit)
	go func() {
		if err := d.bot.Text(msg); err != nil {
			log.Printf("send message error: %s, msg: %s", err, msg)
		}
	}()
}

func (d *DingTalk) SendText(message string) error {
	return d.bot.Text(message)
}
