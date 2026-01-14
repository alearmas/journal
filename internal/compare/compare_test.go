package compare_test

import (
	"alearmas/tradingJournal/internal/compare"
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

func TestSimpleGross_DayCount360(t *testing.T) {
	// 360 convention: 1000 * 10% * 36/360 = 10.00
	got := compare.SimpleGross(D("1000"), D("10"), 36)
	if got.StringFixed(2) != "10.00" {
		t.Fatalf("expected 10.00, got %s", got.StringFixed(2))
	}
}

func TestCompare_Run_NetMath(t *testing.T) {
	out := compare.Run(compare.CompareInput{
		Principal:    D("1000000"),
		Days:         1,
		CaucionTNA:   D("85.5"),
		CaucionFees:  D("50"),
		CaucionTaxes: D("421"),
		PFTNA:        D("80"),
		MMTNA:        D("70"),
	})

	// With 360: gross caucion = 2375.00 -> net = 2375 - 50 - 421 = 1904.00
	if out.CaucionNet.StringFixed(2) != "1904.00" {
		t.Fatalf("expected caucion net 1904.00, got %s", out.CaucionNet.StringFixed(2))
	}
}
