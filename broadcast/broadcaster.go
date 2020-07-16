package broadcast

type Broadcaster interface {
	Broadcast(labels []string, direction, price, amount, total, profit string)
}

const (
	Buy  = "buy"
	Sell = "sell"
)

type Config struct {
	Name    string
	BaseUrl string `json:"baseUrl"`
	Secret  string
}
