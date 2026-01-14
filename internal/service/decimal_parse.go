package service

import "github.com/shopspring/decimal"

func ParseDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}
