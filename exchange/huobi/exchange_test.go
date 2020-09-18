package huobi

import (
	"github.com/xyths/hs/exchange"
	"testing"
)

func TestInterface(t *testing.T) {
	var ex exchange.Exchange
	ex = New("label", "key", "secret", "host")
	_ = ex
}