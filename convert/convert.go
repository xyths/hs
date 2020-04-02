package convert

import "math/big"

func StrToInt64(s string, i64 *int64) {
	if i, ok := big.NewInt(0).SetString(s, 0); ok {
		*i64 = i.Int64()
	}
}
func StrToFloat64(s string) float64 {
	if f, ok := big.NewFloat(0).SetString(s); ok {
		f64, _ := f.Float64()
		return f64
	}
	return 0.0
}
