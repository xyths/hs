package convert

import (
	"math/big"
	"reflect"
	"strconv"
)

func StrToUint64(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

func StrToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func StrToFloat64(s string) float64 {
	if f, ok := big.NewFloat(0).SetString(s); ok {
		f64, _ := f.Float64()
		return f64
	}
	return 0.0
}

func ToFloat64(i interface{}) float64 {
	switch v := reflect.ValueOf(i); v.Kind() {
	case reflect.String:
		if f, err := strconv.ParseFloat(v.String(), 64); err != nil {
			return 0.0
		} else {
			return f
		}
	case reflect.Float64:
		return v.Float()
	default:
		return 0.0
	}
}
