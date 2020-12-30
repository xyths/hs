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

// ResponseReqOrder 是order.query请求返回的主结构
type ResponseReqOrder struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
	// Total 是全部的订单个数，而不是当前批次取到多少
	Total   int64           `json:"total"`
	Records []WsOrderRecord `json:"records"`
}

// WsOrderRecord 在order.query和order.update都相同，可以通用
// 差别是order.update里不含text
type WsOrderRecord struct {
	Id           uint64  `json:"id"`
	Market       string  `json:"market"`
	User         int64   `json:"user"`
	CTime        float64 `json:"ctime"`
	FTime        float64 `json:"ftime"`
	Price        string  `json:"price"`
	Amount       string  `json:"amount"`
	Left         string  `json:"left"`
	DealFee      string  `json:"dealFee"`
	OrderType    int     `json:"orderType"`
	Type         int     `json:"type"`
	FilledAmount string  `json:"filledAmount"`
	FilledTotal  string  `json:"filledTotal"`
	// 文档中有，但实际上却没有该字段
	Text string `json:"text"`

	// 以下字段不在文档中，是调试时看到的字段
	Tif           int64   `json:"tif"`
	MTime         float64 `json:"mtime"`
	Iceberg       string  `json:"iceberg"` // 冰山
	DealFeeRebate string  `json:"deal_fee_rebate"`
	DealPointFee  string  `json:"deal_point_fee"`
	GtDiscount    string  `json:"gt_discount"`
	GtMakerFee    string  `json:"gt_maker_fee"`
	GtTakerFee    string  `json:"gt_taker_fee"`
	DealGtFee     string  `json:"deal_gt_fee"`
}

type UpdateWsBase struct {
	Id     int64       `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}
