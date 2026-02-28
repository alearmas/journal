package sqlite_test

import (
	"alearmas/tradingJournal/internal/adapter/sqlite"
	"alearmas/tradingJournal/internal/domain"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func De(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestSQLiteRepository_AppendAndList(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "journal.db")

	repo, err := sqlite.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("new sqlite repo: %v", err)
	}

	c1 := domain.Caucion{
		ID:            "1",
		Broker:        "Balanz",
		TradeDate:     time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		MaturityDate:  time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
		TermDays:      1,
		Principal:     De("1000.00"),
		TNA:           De("85.5"),
		GrossInterest: De("2.00"),
		Fees:          De("0.10"),
		Taxes:         De("0.20"),
		NetInterest:   De("1.70"),
		Notes:         "x",
		CreatedAt:     time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
	}

	c2 := c1
	c2.ID = "2"
	c2.Notes = "y"

	if err := repo.Append(context.Background(), c1); err != nil {
		t.Fatalf("append c1: %v", err)
	}
	if err := repo.Append(context.Background(), c2); err != nil {
		t.Fatalf("append c2: %v", err)
	}

	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].ID != "1" || items[1].ID != "2" {
		t.Fatalf("unexpected ids order: %s, %s", items[0].ID, items[1].ID)
	}

	if items[0].Principal.StringFixed(2) != "1000.00" {
		t.Fatalf("principal mismatch: %s", items[0].Principal.StringFixed(2))
	}
	if items[0].NetInterest.StringFixed(2) != "1.70" {
		t.Fatalf("net mismatch: %s", items[0].NetInterest.StringFixed(2))
	}
}
