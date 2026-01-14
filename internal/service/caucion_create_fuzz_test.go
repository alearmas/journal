package service_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/service"
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type inMemoryRepo struct {
	items []domain.Caucion
}

func (r *inMemoryRepo) List(ctx context.Context) ([]domain.Caucion, error) {
	return append([]domain.Caucion(nil), r.items...), nil
}

func (r *inMemoryRepo) Append(ctx context.Context, c domain.Caucion) error {
	r.items = append(r.items, c)
	return nil
}

func decFromParts(unscaled int64, scale int32) decimal.Decimal {
	// Example: unscaled=12345, scale=2 => 123.45
	return decimal.New(unscaled, -scale)
}

func FuzzCaucionService_Create(f *testing.F) {
	// Seeds: “normal” + edge-ish
	f.Add(int64(100000000), int32(2), int64(855), int32(1), int32(1), int64(5000), int64(42100)) // 1,000,000.00 | 85.5 | 1 day | 50.00 | 421.00
	f.Add(int64(1), int32(2), int64(1), int32(2), int32(1), int64(0), int64(0))                  // 0.01 | 0.01 | 1 day
	f.Add(int64(-1000), int32(2), int64(855), int32(1), int32(7), int64(0), int64(0))            // negative principal (should error)
	f.Add(int64(100000), int32(2), int64(-10), int32(1), int32(7), int64(0), int64(0))           // negative tna (should error)
	f.Add(int64(100000), int32(2), int64(855), int32(1), int32(0), int64(0), int64(0))           // zero days (should error)

	f.Fuzz(func(t *testing.T,
		principalUnscaled int64, principalScale int32,
		tnaUnscaled int64, tnaScale int32,
		termDays int32,
		feesCents int64,
		taxesCents int64,
	) {
		// Keep scale in a sane range to avoid absurd decimals that slow tests.
		if principalScale < 0 {
			principalScale = -principalScale
		}
		if principalScale > 6 {
			principalScale = 6
		}
		if tnaScale < 0 {
			tnaScale = -tnaScale
		}
		if tnaScale > 6 {
			tnaScale = 6
		}

		// Bound termDays to avoid huge date jumps and slow fuzzing.
		// Allow negatives/zero to test validation paths.
		if termDays > 3650 {
			termDays = 3650
		}
		if termDays < -3650 {
			termDays = -3650
		}

		repo := &inMemoryRepo{}
		svc := service.NewCaucionService(repo)

		tradeDate := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

		in := service.CreateCaucionInput{
			Broker:    "Balanz",
			TradeDate: tradeDate,
			TermDays:  int(termDays),

			Principal: decFromParts(principalUnscaled, principalScale),
			TNA:       decFromParts(tnaUnscaled, tnaScale),

			Fees:  decimal.NewFromInt(feesCents).Div(decimal.NewFromInt(100)),
			Taxes: decimal.NewFromInt(taxesCents).Div(decimal.NewFromInt(100)),

			Notes: "fuzz",
		}

		// If this panics, fuzz test fails automatically.
		c, err := svc.Create(context.Background(), in)

		// Success path invariants
		if err == nil {
			// maturity = tradeDate + termDays
			wantMaturity := tradeDate.AddDate(0, 0, int(termDays))
			if !c.MaturityDate.Equal(wantMaturity) {
				t.Fatalf("maturity mismatch: got %s want %s", c.MaturityDate, wantMaturity)
			}

			// net = gross - fees - taxes (rounded to 2)
			wantNet := c.GrossInterest.Sub(c.Fees).Sub(c.Taxes).Round(2)
			if !c.NetInterest.Equal(wantNet) {
				t.Fatalf("net mismatch: got %s want %s", c.NetInterest.StringFixed(2), wantNet.StringFixed(2))
			}

			// Stored once
			items, _ := repo.List(context.Background())
			if len(items) != 1 {
				t.Fatalf("expected 1 stored item, got %d", len(items))
			}
		}
	})
}
