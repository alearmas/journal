package report_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/report"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func D(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestFilterByMonth(t *testing.T) {
	items := []domain.Caucion{
		{ID: "1", TradeDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
		{ID: "2", TradeDate: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)},
		{ID: "3", TradeDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	got, err := report.FilterByMonth(items, "2026-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "2" {
		t.Fatalf("unexpected ids: %v, %v", got[0].ID, got[1].ID)
	}
}

func TestSummarizeMonth(t *testing.T) {
	items := []domain.Caucion{
		{
			Principal:     D("1000.00"),
			TNA:           D("80"),
			GrossInterest: D("10.00"),
			Fees:          D("1.00"),
			Taxes:         D("2.00"),
			NetInterest:   D("7.00"),
		},
		{
			Principal:     D("3000.00"),
			TNA:           D("100"),
			GrossInterest: D("30.00"),
			Fees:          D("2.00"),
			Taxes:         D("3.00"),
			NetInterest:   D("25.00"),
		},
	}

	s := report.SummarizeMonth(items, "2026-01")

	if s.Count != 2 {
		t.Fatalf("expected count 2, got %d", s.Count)
	}
	if s.TotalPrincipal.StringFixed(2) != "4000.00" {
		t.Fatalf("principal total mismatch: %s", s.TotalPrincipal.StringFixed(2))
	}
	if s.TotalNet.StringFixed(2) != "32.00" {
		t.Fatalf("net total mismatch: %s", s.TotalNet.StringFixed(2))
	}

	// WeightedAvgTNA = (1000*80 + 3000*100) / 4000 = 95
	if s.WeightedAvgTNA.String() != "95" && s.WeightedAvgTNA.StringFixed(4) != "95.0000" {
		t.Fatalf("weighted avg TNA mismatch: %s", s.WeightedAvgTNA.String())
	}
}
