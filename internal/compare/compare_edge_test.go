package compare_test

import (
	"alearmas/tradingJournal/internal/compare"
	"testing"

	"github.com/shopspring/decimal"
)

func TestSimpleGross_ZeroDays(t *testing.T) {
	got := compare.SimpleGross(D("1000"), D("10"), 0)
	if !got.Equal(decimal.Zero) {
		t.Fatalf("expected 0, got %s", got.String())
	}
}

func TestSimpleGross_ZeroPrincipal(t *testing.T) {
	got := compare.SimpleGross(decimal.Zero, D("10"), 30)
	if !got.Equal(decimal.Zero) {
		t.Fatalf("expected 0, got %s", got.String())
	}
}

func TestCompare_Run_AllZero(t *testing.T) {
	out := compare.Run(compare.CompareInput{
		Principal:    decimal.Zero,
		Days:         0,
		CaucionTNA:   decimal.Zero,
		CaucionFees:  decimal.Zero,
		CaucionTaxes: decimal.Zero,
		PFTNA:        decimal.Zero,
		MMTNA:        decimal.Zero,
	})

	if !out.CaucionNet.Equal(decimal.Zero) {
		t.Fatalf("expected zero caucion net, got %s", out.CaucionNet.String())
	}
	if !out.PFGross.Equal(decimal.Zero) {
		t.Fatalf("expected zero PF gross, got %s", out.PFGross.String())
	}
	if !out.MMGross.Equal(decimal.Zero) {
		t.Fatalf("expected zero MM gross, got %s", out.MMGross.String())
	}
}

func TestCompare_Run_HighFees(t *testing.T) {
	// When fees+taxes > gross, net should be negative
	out := compare.Run(compare.CompareInput{
		Principal:    D("1000"),
		Days:         1,
		CaucionTNA:   D("10"),
		CaucionFees:  D("100"),
		CaucionTaxes: D("100"),
		PFTNA:        D("10"),
		MMTNA:        D("10"),
	})

	if out.CaucionNet.Cmp(decimal.Zero) >= 0 {
		t.Fatalf("expected negative net when fees exceed gross, got %s", out.CaucionNet.String())
	}
}
