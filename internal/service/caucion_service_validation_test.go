package service_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/service"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// failingRepo always returns an error on Append.
type failingRepo struct {
	err error
}

func (r *failingRepo) List(_ context.Context) ([]domain.Caucion, error) {
	return nil, nil
}

func (r *failingRepo) Append(_ context.Context, _ domain.Caucion) error {
	return r.err
}

func validInput() service.CreateCaucionInput {
	return service.CreateCaucionInput{
		Broker:    "Balanz",
		TradeDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		TermDays:  1,
		Principal: decimal.NewFromInt(1000),
		TNA:       decimal.NewFromInt(85),
		Fees:      decimal.Zero,
		Taxes:     decimal.Zero,
		Notes:     "test",
	}
}

func TestCreate_ValidationErrors(t *testing.T) {
	repo := &inMemoryRepository{}
	svc := service.NewCaucionService(repo)
	ctx := context.Background()

	tests := []struct {
		name  string
		mod   func(*service.CreateCaucionInput)
		field string
	}{
		{"zero termDays", func(in *service.CreateCaucionInput) { in.TermDays = 0 }, "termDays"},
		{"negative termDays", func(in *service.CreateCaucionInput) { in.TermDays = -5 }, "termDays"},
		{"zero principal", func(in *service.CreateCaucionInput) { in.Principal = decimal.Zero }, "principal"},
		{"negative principal", func(in *service.CreateCaucionInput) { in.Principal = decimal.NewFromInt(-100) }, "principal"},
		{"zero TNA", func(in *service.CreateCaucionInput) { in.TNA = decimal.Zero }, "TNA"},
		{"negative TNA", func(in *service.CreateCaucionInput) { in.TNA = decimal.NewFromInt(-10) }, "TNA"},
		{"negative fees", func(in *service.CreateCaucionInput) { in.Fees = decimal.NewFromInt(-1) }, "fees"},
		{"negative taxes", func(in *service.CreateCaucionInput) { in.Taxes = decimal.NewFromInt(-1) }, "taxes"},
		// validInput: principal=1000, tna=85, term=1 → gross ≈ 2.36
		// Setting fees=100 makes fees+taxes (100) > gross (2.36)
		{"fees exceed gross", func(in *service.CreateCaucionInput) { in.Fees = decimal.NewFromInt(100) }, "fees+taxes"},
		{"taxes exceed gross", func(in *service.CreateCaucionInput) { in.Taxes = decimal.NewFromInt(100) }, "fees+taxes"},
		{"fees+taxes exceed gross", func(in *service.CreateCaucionInput) {
			in.Fees = D("1.5")
			in.Taxes = D("1.5") // 3.00 > 2.36
		}, "fees+taxes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.mod(&in)

			_, err := svc.Create(ctx, in)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var ve *domain.ErrValidation
			if !errors.As(err, &ve) {
				t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
			}
			if ve.Field != tt.field {
				t.Fatalf("expected field %q, got %q", tt.field, ve.Field)
			}
		})
	}
}

func TestCreate_DefaultBroker(t *testing.T) {
	repo := &inMemoryRepository{}
	svc := service.NewCaucionService(repo)

	in := validInput()
	in.Broker = ""

	c, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Broker != "Balanz" {
		t.Fatalf("expected default broker Balanz, got %s", c.Broker)
	}
}

func TestCreate_DefaultTradeDate(t *testing.T) {
	repo := &inMemoryRepository{}
	svc := service.NewCaucionService(repo)

	in := validInput()
	in.TradeDate = time.Time{} // zero value

	c, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.TradeDate.IsZero() {
		t.Fatal("expected non-zero trade date")
	}
}

func TestCreate_RepoError(t *testing.T) {
	repoErr := fmt.Errorf("disk full")
	repo := &failingRepo{err: repoErr}
	svc := service.NewCaucionService(repo)

	in := validInput()
	_, err := svc.Create(context.Background(), in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}

func TestCreate_MaturityCalculation(t *testing.T) {
	repo := &inMemoryRepository{}
	svc := service.NewCaucionService(repo)

	in := validInput()
	in.TermDays = 30

	c, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := in.TradeDate.AddDate(0, 0, 30)
	if !c.MaturityDate.Equal(want) {
		t.Fatalf("maturity: got %s, want %s", c.MaturityDate, want)
	}
}

func TestCreate_NetInterestCalculation(t *testing.T) {
	repo := &inMemoryRepository{}
	svc := service.NewCaucionService(repo)

	in := validInput()
	in.Principal = decimal.NewFromInt(1000000)
	in.TNA = D("85.5")
	in.TermDays = 1
	in.Fees = D("50")
	in.Taxes = D("421")

	c, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// gross = 1000000 * 85.5/100 * 1/360 = 2375.00
	// net = 2375 - 50 - 421 = 1904.00
	if c.NetInterest.StringFixed(2) != "1904.00" {
		t.Fatalf("net interest: got %s, want 1904.00", c.NetInterest.StringFixed(2))
	}
}
