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
	//c := dataframe.New(
	//	series.New([]int64{}, series.Int, "Timestamp"),
	//	series.New([]float64{}, series.Float, "Open"),
	//	series.New([]float64{}, series.Float, "High"),
	//	series.New([]float64{}, series.Float, "Low"),
	//	series.New([]float64{}, series.Float, "Close"),
	//	series.New([]float64{}, series.Float, "Volume"),
	//)
	return Candle{Capacity: capacity}
}

func (c Candle) Length() int {
	return len(c.Timestamp)
}

func (c *Candle) Append(ticker Ticker) {
	if c.Capacity == 0 {
		return
	}
	pos := c.Length()
	if c.Length() >= c.Capacity {
		pos = c.Capacity - 1
	}
	c.Timestamp = append([]int64{ticker.Timestamp}, c.Timestamp[0:pos]...)
	c.Open = append([]float64{ticker.Open}, c.Open[0:pos]...)
	c.High = append([]float64{ticker.High}, c.High[0:pos]...)
	c.Low = append([]float64{ticker.Low}, c.Low[0:pos]...)
	c.Close = append([]float64{ticker.Close}, c.Close[0:pos]...)
	c.Volume = append([]float64{ticker.Volume}, c.Volume[0:pos]...)
}

func (c *Candle) Add(other Candle) {
	if c.Capacity == 0 {
		return
	}
	l1 := c.Length()
	l2 := other.Length()

	if l1+l2 <= c.Capacity {
		c.Timestamp = append(other.Timestamp, c.Timestamp...)
		c.Open = append(other.Open, c.Open...)
		c.High = append(other.High, c.High...)
		c.Low = append(other.Low, c.Low...)
		c.Close = append(other.Close, c.Close...)
		c.Volume = append(other.Volume, c.Volume...)
	} else if l2 <= c.Capacity {
		pos := c.Capacity - l2
		c.Timestamp = append(other.Timestamp, c.Timestamp[0:pos]...)
		c.Open = append(other.Open, c.Open[0:pos]...)
		c.High = append(other.High, c.High[0:pos]...)
		c.Low = append(other.Low, c.Low[0:pos]...)
		c.Close = append(other.Close, c.Close[0:pos]...)
		c.Volume = append(other.Volume, c.Volume[0:pos]...)
	} else {
		c.Timestamp = other.Timestamp[0:c.Capacity]
		c.Open = other.Open[0:c.Capacity]
		c.High = other.High[0:c.Capacity]
		c.Low = other.Low[0:c.Capacity]
		c.Close = other.Close[0:c.Capacity]
		c.Volume = other.Volume[0:c.Capacity]
	}
}
