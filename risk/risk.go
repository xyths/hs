package risk

import "github.com/shopspring/decimal"

func SpotRisk(buy, stop, amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(buy.Sub(stop))
}

func ContractRisk(buy, stop decimal.Decimal, unit, number int) decimal.Decimal {
	return buy.Sub(stop).Abs().Mul(decimal.NewFromInt(int64(unit * number)))
}

type Spot struct {
	PricePrecision  int32
	AmountPrecision int32
}

// Quota is max amount when place order
func (s Spot) Quota(buy, stop, total decimal.Decimal) decimal.Decimal {
	buy = buy.Round(s.PricePrecision)
	stop = stop.Round(s.PricePrecision)
	return total.DivRound(buy.Sub(stop), s.AmountPrecision)
}

// QuotaAtr is Quota by atr
func (s Spot) QuotaAtr(atr float64, total decimal.Decimal) decimal.Decimal {
	return total.DivRound(decimal.NewFromFloat(atr).Round(s.PricePrecision), s.AmountPrecision)
}
