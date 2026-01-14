package service_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/service"
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type inMemoryRepository struct {
	items []domain.Caucion
}

func (r *inMemoryRepository) List(ctx context.Context) ([]domain.Caucion, error) {
	return append([]domain.Caucion(nil), r.items...), nil
}

func (r *inMemoryRepository) Append(ctx context.Context, c domain.Caucion) error {
	r.items = append(r.items, c)
	return nil
}

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
	// gross = 1000000 * 0.855 * (1/360) = 2375.00
	gross := service.ComputeGrossInterest(D("1000000"), D("85.5"), 1)
	if gross.StringFixed(2) != "2375.00" {
		t.Fatalf("expected 2375.00, got %s", gross.StringFixed(2))
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

func TestCaucionService_List(t *testing.T) {
	repo := &inMemoryRepository{
		items: []domain.Caucion{
			{
				ID:            "1",
				Broker:        "Balanz",
				TradeDate:     time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
				MaturityDate:  time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
				TermDays:      1,
				Principal:     D("1000.00"),
				TNA:           D("85.5"),
				GrossInterest: D("2.00"),
				Fees:          D("0.10"),
				Taxes:         D("0.20"),
				NetInterest:   D("1.70"),
				Notes:         "x",
				CreatedAt:     time.Now(),
			},
		},
	}

	svc := service.NewCaucionService(repo)

	items, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "1" {
		t.Fatalf("unexpected ID: %s", items[0].ID)
	}
}
