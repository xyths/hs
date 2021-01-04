package gateio

import (
	"github.com/shopspring/decimal"
	"testing"
)

func Test_TakeAsks(t *testing.T) {
	var tests = []struct {
		Asks         []Quote
		Total        decimal.Decimal
		ExpectPrice  decimal.Decimal
		ExpectAmount decimal.Decimal
	}{
		{[]Quote{{1.2, 1.0}, {1.0, 1.0}}, decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.0)},
		{[]Quote{{1.2, 1.0}, {1.0, 1.0}}, decimal.NewFromFloat(2.2), decimal.NewFromFloat(1.2), decimal.NewFromFloat(1.0)},
	}
	for i, tt := range tests {
		p, a := takeAsks(tt.Asks, tt.Total, 2, 2)
		if !p.Equal(tt.ExpectPrice) {
			t.Errorf("[%d] price expect %s, actual %s", i, tt.ExpectPrice, p)
		}
		if !a.Equal(tt.ExpectAmount) {
			t.Errorf("[%d] amount expect %s, actual %s", i, tt.ExpectAmount, a)
		}
	}
}

func Test_TakeBids(t *testing.T) {
	var tests = []struct {
		Bids        []Quote
		Amount      decimal.Decimal
		ExpectPrice decimal.Decimal
	}{
		{[]Quote{{1.0, 1.0}, {0.9, 1.0}}, decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.0)},
		{[]Quote{{1.0, 1.0}, {0.9, 1.0}}, decimal.NewFromFloat(2.0), decimal.NewFromFloat(0.9)},
	}
	for i, tt := range tests {
		p := takeBids(tt.Bids, tt.Amount, 2, 2)
		if !p.Equal(tt.ExpectPrice) {
			t.Errorf("[%d] price expect %s, actual %s", i, tt.ExpectPrice, p)
		}
	}
}
