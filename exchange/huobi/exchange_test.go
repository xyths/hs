package huobi

import (
	"github.com/xyths/hs"
	"testing"
)

func TestInterface(t *testing.T) {
	var ex hs.Exchange
	ex = New("label", "key", "secret", "host")
	_ = ex
}
