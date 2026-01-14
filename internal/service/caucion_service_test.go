package service_test

import (
	"alearmas/tradingJournal/internal/service"
	"testing"

	"github.com/shopspring/decimal"
)

func D(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestComputeGrossInterest_Rounds2(t *testing.T) {
	// Example:
	// principal=1,000,000 ; tna=85.5 ; term=1 day
	// gross = 1000000 * 0.855 * (1/365) = 2342.465753... -> 2342.47
	gross := service.ComputeGrossInterest(D("1000000"), D("85.5"), 1)
	if gross.StringFixed(2) != "2342.47" {
		t.Fatalf("expected 2342.47, got %s", gross.StringFixed(2))
	}
}

func TestComputeGrossInterest_Deterministic(t *testing.T) {
	// Ensure no float drift: same inputs => exact same output
	a := service.ComputeGrossInterest(D("123456.78"), D("97.25"), 7)
	b := service.ComputeGrossInterest(D("123456.78"), D("97.25"), 7)
	if !a.Equal(b) {
		t.Fatalf("expected deterministic result, got %s vs %s", a.String(), b.String())
	}
}
