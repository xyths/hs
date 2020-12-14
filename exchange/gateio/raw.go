package gateio

type ResponseWsBase struct {
	Id     int64       `json:"id"`
	Error  interface{} `json:"error"`
	Result interface{} `json:"result"`
}

type ResponseWsTicker struct {
	Period      int64  `json:"period"`
	Open        string `json:"open"`
	High        string `json:"high"`
	Low         string `json:"low"`
	Close       string `json:"close"`
	Last        string `json:"last"`
	Change      string `json:"change"`
	QuoteVolume string `json:"quoteVolume"`
	BaseVolume  string `json:"baseVolume"`
}

type ResponseWsKline struct {
	Time   int64  `json:"time"`
	Open   string `json:"open"`
	High   string `json:"highest"`
	Low    string `json:"lowest"`
	Close  string `json:"close"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Symbol string `json:"market_name"`
}

type UpdateWsBase struct {
	Id     int64       `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}
