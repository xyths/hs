package hs

import "testing"

func TestCandle_Append(t *testing.T) {
	candles := []Candle{
		NewCandle(0),
		NewCandle(1),
		NewCandle(2),
		NewCandle(3),
	}
	tickers := []Ticker{
		{1, 1, 1, 1, 1, 1,},
		{2, 2, 2, 2, 2, 2,},
	}
	for j, c := range candles {
		for i, ticker := range tickers {
			c.Append(ticker)

			expect := i + 1
			if expect > c.Capacity {
				expect = c.Capacity
			}
			if c.Length() != expect {
				t.Errorf("candles %d expect length: %d, got: %d", j, expect, c.Length())
			}
		}
	}
}

func TestCandle_Add(t *testing.T) {
	tickers := []Ticker{
		{0, 0, 0, 0, 0, 0,},
		{1, 1, 1, 1, 1, 1,},
		{2, 2, 2, 2, 2, 2,},
		{3, 3, 3, 3, 3, 3,},
		{2, 4, 4, 4, 4, 4,},
	}
	c1 := NewCandle(0)
	c2 := NewCandle(0)
	c3 := NewCandle(3)
	c3.Append(tickers[1])
	c4 := NewCandle(1)
	c4.Append(tickers[2])
	c5 := NewCandle(2)
	c5.Append(tickers[3])

	t.Run("empty add empty", func(t *testing.T) {
		c1.Add(c2)
		if c1.Length() != 0 {
			t.Errorf("length expect 0, got %d", c1.Length())
		}
	})

	t.Run("L1 + L2 < Cap", func(t *testing.T) {
		c3.Add(c1)
		if c3.Length() != 1 {
			t.Errorf("length expect 1, actual %d", c3.Length())
		}

		c3.Add(c4)
		if c3.Length() != 2 {
			t.Errorf("length expect 2, actual %d", c3.Length())
		}
		if c3.Timestamp[0] != 2 {
			t.Errorf("timestamp expect 2, actual %d", c3.Timestamp[0])
		}
	})

	t.Run("L1 + L2 == Cap", func(t *testing.T) {
		c3.Add(c5)
		if c3.Length() != 3 {
			t.Errorf("length expect 3, actual %d", c3.Length())
		}
		if c3.Timestamp[0] != 3 {
			t.Errorf("timestamp expect 3, actual %d", c3.Timestamp[0])
		}
	})

	t.Run("L1 + L2 >= Cap && L2 <= Cap", func(t *testing.T) {
		c1.Capacity = 2
		c1.Append(tickers[1])
		c2.Capacity = 2
		c2.Append(tickers[1])
		c2.Append(tickers[2])
		c2.Add(c1)
		if c2.Length() != c2.Capacity {
			t.Errorf("length expect %d, actual %d", c2.Capacity, c2.Length())
		}
		if c2.Timestamp[0] != 1 {
			t.Errorf("timestamp expect 1, actual %d", c2.Timestamp[0])
		}

		c1.Add(c2)
		if c1.Length() != c1.Capacity {
			t.Errorf("length expect %d, actual %d", c1.Capacity, c1.Length())
		}
		if c1.Timestamp[1] != 2 {
			t.Errorf("timestamp expect 2, actual %d", c1.Timestamp[1])
		}
	})

	t.Run("L1 + L2 >= Cap && L2 > Cap", func(t *testing.T) {
		c4.Add(c3)
		if c4.Length() != c4.Capacity {
			t.Errorf("length expect %d, actual %d", c4.Capacity, c4.Length())
		}
		if c4.Timestamp[0] != 3 {
			t.Errorf("timestamp expect 3, actual %d", c4.Timestamp[0])
		}
	})
}

func TestMergeTick(t *testing.T) {
	var tests = [][15]float64{
		{
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 2,
		},
		{
			1, 1, 1, 0, 2,
			1, 2, 2, 2, 1,
			1, 2, 1, 2, 3,
		},
		{
			1, 3, 2, 0, 1,
			1, 2, 1, 0, 2,
			1, 3, 1, 0, 3,
		},
	}
	for i, tt := range tests {
		o, h, l, c, v := MergeTick(tt[0], tt[1], tt[2], tt[3],
			tt[4], tt[5], tt[6], tt[7], tt[8], tt[9])
		if o != tt[10] || h != tt[11] || l != tt[12] || c != tt[13] || v != tt[14] {
			t.Logf("%d: wrong, want [%f %f %f %f %f], got [%f %f %f %f %f]", i, o, h, l, c, v,
				tt[10], tt[11], tt[12], tt[13], tt[14])
		}
	}
}
