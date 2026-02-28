package report_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/report"
	"testing"

	"github.com/shopspring/decimal"
)

func TestFilterByMonth_InvalidFormat(t *testing.T) {
	_, err := report.FilterByMonth(nil, "not-a-month")
	if err == nil {
		t.Fatal("expected error for invalid month format")
	}
}

func TestFilterByMonth_Empty(t *testing.T) {
	got, err := report.FilterByMonth(nil, "2026-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}

func TestSummarizeMonth_Empty(t *testing.T) {
	s := report.SummarizeMonth(nil, "2026-01")
	if s.Count != 0 {
		t.Fatalf("expected count 0, got %d", s.Count)
	}
	if !s.WeightedAvgTNA.Equal(decimal.Zero) {
		t.Fatalf("expected zero weighted TNA, got %s", s.WeightedAvgTNA.String())
	}
}

func TestSummarizeMonth_SingleItem(t *testing.T) {
	items := []domain.Caucion{
		{
			Principal:     D("5000"),
			TNA:           D("90"),
			GrossInterest: D("12.50"),
			Fees:          D("1.00"),
			Taxes:         D("2.00"),
			NetInterest:   D("9.50"),
		},
	}

	s := report.SummarizeMonth(items, "2026-06")
	if s.Count != 1 {
		t.Fatalf("expected count 1, got %d", s.Count)
	}
	if s.TotalPrincipal.StringFixed(2) != "5000.00" {
		t.Fatalf("principal: got %s", s.TotalPrincipal.StringFixed(2))
	}
	// Weighted avg TNA with single item = TNA itself
	if s.WeightedAvgTNA.String() != "90" {
		t.Fatalf("weighted TNA: got %s, want 90", s.WeightedAvgTNA.String())
	}
}
