package export_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/export"
	"os"
	"path/filepath"
	"strings"
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

func TestWriteCaucionesCSV_WritesHeaderAndRow(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "cauciones.csv")

	items := []domain.Caucion{
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
			CreatedAt:     time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
		},
	}

	if err := export.WriteCaucionesCSV(out, items); err != nil {
		t.Fatalf("export error: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	s := string(b)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}

	if !strings.HasPrefix(lines[0], "id,broker,trade_date") {
		t.Fatalf("unexpected header: %s", lines[0])
	}
	if !strings.Contains(lines[1], "1,Balanz,2026-01-10,2026-01-11,1,1000.00,85.5,2.00,0.10,0.20,1.70,x,") {
		t.Fatalf("unexpected row: %s", lines[1])
	}
}
