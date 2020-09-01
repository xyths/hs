package gateio

import (
	"github.com/xyths/hs"
	"testing"
)

// test the interface requirement
func TestInterface(t *testing.T) {
	var ex hs.RestAPIExchange
	ex = New("key", "secret", "host")
	_ = ex
}
