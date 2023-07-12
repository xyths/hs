package convert

import "testing"

func TestShortAddress(t *testing.T) {
	tests := []string{
		"0x9566a524fd5d4f67514151e6c74126ec7deb39ac",
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
	}
	for _, tt := range tests {
		t.Logf("%s => %s", tt, ShortAddress(tt))
	}
}
