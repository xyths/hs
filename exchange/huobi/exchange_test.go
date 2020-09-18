package huobi

import (
	"github.com/stretchr/testify/require"
	"github.com/xyths/hs/exchange"
	"testing"
)

func TestInterface(t *testing.T) {
	var ex exchange.Exchange
	ex, err := New("label", "key", "secret", "host")
	require.NoError(t, err)
	_ = ex
}
