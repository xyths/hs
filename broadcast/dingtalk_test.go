package broadcast

import (
	"os"
	"testing"
	"time"
)

func TestDingTalk_Broadcast(t *testing.T) {
	baseUrl := os.Getenv("BASE_URL")
	secret := os.Getenv("SECRET")
	dt := NewDingTalk(Config{
		BaseUrl: baseUrl,
		Secret:  secret,
	})
	dt.Broadcast([]string{
		"Gate", "tangzhu01",
	}, "BTC_USDT", "buy", "0.19", "0.01", "123", "2.1")
	time.Sleep(time.Minute)
}
