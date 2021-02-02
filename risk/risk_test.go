package risk

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestSpotRisk(t *testing.T) {
	tests := []struct {
		Buy    decimal.Decimal
		Stop   decimal.Decimal
		Amount decimal.Decimal
		Risk   decimal.Decimal
	}{
		{decimal.NewFromInt(2), decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(1)},
	}
	for i, tt := range tests {
		risk := SpotRisk(tt.Buy, tt.Stop, tt.Amount)
		if !risk.Equal(tt.Risk) {
			t.Errorf("[%d] expect %s, got %s", i, tt.Risk, risk)
		}
	}
}

func TestContractRisk(t *testing.T) {
	tests := []struct {
		Buy    decimal.Decimal
		Stop   decimal.Decimal
		Unit   int
		Number int
		Risk   decimal.Decimal
	}{
		{decimal.NewFromInt(2), decimal.NewFromInt(1), 1, 1, decimal.NewFromInt(1)},
		// 该例子来自《海龟交易法则(珍藏版)》第三章海龟培训课程第3节海龟的优势
		// 假设有一笔黄金交易，买入价是350美元，止损价是320美元，共买入了10份合约，
		// 那么这笔交易的风险水平等于：
		// 买入价与止损价的差异（30美元）乘以合约数量（10份合约）再乘以合约本身的大小（每份合约100盎司黄金），
		// 也就是30000美元。
		{decimal.NewFromInt(350), decimal.NewFromInt(320), 100, 10, decimal.NewFromInt(30000)},
		{decimal.NewFromInt(1810), decimal.NewFromInt(1816), 50, 1, decimal.NewFromInt(300)},
	}
	for i, tt := range tests {
		risk := ContractRisk(tt.Buy, tt.Stop, tt.Unit, tt.Number)
		if !risk.Equal(tt.Risk) {
			t.Errorf("[%d] expect %s, got %s", i, tt.Risk, risk)
		}
	}
}

func TestSpot_Quota(t *testing.T) {
	type testcase struct {
		Buy, Stop, Total, Quota decimal.Decimal
	}
	t.Run("stock spot", func(t *testing.T) {
		tests := []testcase{
			// 该例子来自《以交易为生》第9章风险管理第50节 2%法则
			// 假定你决定按40美元的价格买入股票，止损线设在38美元。
			// 这意味着你每股要承担2美元的风险。
			// 你总的可承受风险为1000美元，除以每股2美元，得到你可以交易不超过500股。
			{decimal.NewFromFloat(40), decimal.NewFromFloat(38), decimal.NewFromFloat(1000), decimal.NewFromFloat(500)},
		}
		spot := Spot{PricePrecision: 0, AmountPrecision: 0}
		for i, tt := range tests {
			quota := spot.Quota(tt.Buy, tt.Stop, tt.Total)
			if !quota.Equal(tt.Quota) {
				t.Errorf("[%d] expect %s, got %s", i, tt.Quota, quota)
			}
		}
	})
	t.Run("crypto spot", func(t *testing.T) {
		tests := []testcase{
			{decimal.NewFromFloat(40000), decimal.NewFromFloat(38000), decimal.NewFromFloat(10000), decimal.NewFromFloat(5.0)},
		}
		spot := Spot{PricePrecision: 1, AmountPrecision: 1}
		for i, tt := range tests {
			quota := spot.Quota(tt.Buy, tt.Stop, tt.Total)
			if !quota.Equal(tt.Quota) {
				t.Errorf("[%d] expect %s, got %s", i, tt.Quota, quota)
			}
		}
	})
}

func TestSpot_QuotaAtr(t *testing.T) {
	type testcase struct {
		Atr          float64
		Total, Quota decimal.Decimal
	}
	t.Run("stock", func(t *testing.T) {
		tests := []testcase{
			{2.0, decimal.NewFromFloat(1000), decimal.NewFromFloat(500)},
		}
		spot := Spot{PricePrecision: 0, AmountPrecision: 0}
		for i, tt := range tests {
			quota := spot.QuotaAtr(tt.Atr, tt.Total)
			if !quota.Equal(tt.Quota) {
				t.Errorf("[%d] expect %s, got %s", i, tt.Quota, quota)
			}
		}
	})
	t.Run("crypto", func(t *testing.T) {
		tests := []testcase{
			{2000, decimal.NewFromFloat(10000), decimal.NewFromFloat(5.0)},
		}
		spot := Spot{PricePrecision: 1, AmountPrecision: 1}
		for i, tt := range tests {
			quota := spot.QuotaAtr(tt.Atr, tt.Total)
			if !quota.Equal(tt.Quota) {
				t.Errorf("[%d] expect %s, got %s", i, tt.Quota, quota)
			}
		}
	})
}
