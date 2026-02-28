package filejson_test

import (
	"alearmas/tradingJournal/internal/adapter/filejson"
	"alearmas/tradingJournal/internal/domain"
	"context"
	"path/filepath"
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

func TestFileJSONRepository_AppendAndList(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cauciones.json")

	repo := filejson.NewFileJSONRepository(path)

	c1 := domain.Caucion{
		ID: "1", Broker: "Balanz",
		TradeDate:    time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		MaturityDate: time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
		TermDays:     1,
		Principal:    D("1000"), TNA: D("85.5"),
		GrossInterest: D("2.34"), Fees: D("0.10"), Taxes: D("0.20"), NetInterest: D("2.04"),
		Notes: "a", CreatedAt: time.Now(),
	}

	c2 := c1
	c2.ID = "2"
	c2.Notes = "b"

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
		t.Fatalf("unexpected order/ids: %+v", items)
	}
}
