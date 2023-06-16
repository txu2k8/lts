package utils

import (
	"github.com/shopspring/decimal"
)

func Zfill(str string, width int) string {
	if len(str) >= width {
		return str
	}
	padding := make([]byte, width-len(str))
	for i := range padding {
		padding[i] = '0'
	}
	return string(padding) + str
}

func Decimal(value float64, exp int32) float64 {
	var d decimal.Decimal
	d = decimal.NewFromFloatWithExponent(value, -2).Round(exp)
	v, _ := d.Float64()
	return v
}
