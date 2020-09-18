package gateio

import (
	"github.com/xyths/hs/exchange"
	"testing"
)

// test the interface requirement
func TestInterface(t *testing.T) {
	var ex exchange.RestAPIExchange
	ex = New("key", "secret", "host")
	_ = ex
}
