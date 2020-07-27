package hs

type Ticker struct {
	Timestamp int64 // unix timestamp in seconds
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type Candle struct {
	Capacity  int
	Timestamp []int64 // unix timestamp in seconds
	Open      []float64
	High      []float64
	Low       []float64
	Close     []float64
	Volume    []float64
}

func NewCandle(capacity int) Candle {
	return Candle{Capacity: capacity}
}

func (c Candle) Length() int {
	return len(c.Timestamp)
}

func (c *Candle) Append(ticker Ticker) {
	if c.Capacity == 0 {
		return
	}
	c.Timestamp = append(c.Timestamp, ticker.Timestamp)
	c.Open = append(c.Open, ticker.Open)
	c.High = append(c.High, ticker.High)
	c.Low = append(c.Low, ticker.Low)
	c.Close = append(c.Close, ticker.Close)
	c.Volume = append(c.Volume, ticker.Volume)

	c.Truncate()
}

func (c *Candle) Add(other Candle) {
	if c.Capacity == 0 {
		return
	}
	if c.Length() >= 1 && other.Length() >= 1 && other.Timestamp[0] == c.Timestamp[c.Length()-1] {
		// remove the duplicate item
		other.Timestamp = other.Timestamp[1:]
		other.Open = other.Open[1:]
		other.High = other.High[1:]
		other.Low = other.Low[1:]
		other.Close = other.Close[1:]
		other.Volume = other.Volume[1:]
	}

	c.Timestamp = append(c.Timestamp, other.Timestamp...)
	c.Open = append(c.Open, other.Open...)
	c.High = append(c.High, other.High...)
	c.Low = append(c.Low, other.Low...)
	c.Close = append(c.Close, other.Close...)
	c.Volume = append(c.Volume, other.Volume...)

	c.Truncate()
}

func (c *Candle) Truncate() {
	if c.Length() <= c.Capacity {
		return
	}
	pos := c.Length() - c.Capacity
	c.Timestamp = c.Timestamp[pos:]
	c.Open = c.Open[pos:]
	c.High = c.High[pos:]
	c.Low = c.Low[pos:]
	c.Close = c.Close[pos:]
	c.Volume = c.Volume[pos:]
}
