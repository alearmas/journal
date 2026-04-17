package sqlite_test

import (
	"alearmas/tradingJournal/internal/adapter/sqlite"
	"alearmas/tradingJournal/internal/domain"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteRepository_DuplicateIDError(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "dup.db")

	repo, err := sqlite.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("new sqlite repo: %v", err)
	}

	c := domain.Caucion{
		ID:            "dup-1",
		Broker:        "Test",
		TradeDate:     time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		MaturityDate:  time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
		TermDays:      1,
		Principal:     De("1000"),
		TNA:           De("85"),
		GrossInterest: De("2"),
		Fees:          De("0"),
		Taxes:         De("0"),
		NetInterest:   De("2"),
		Notes:         "x",
		CreatedAt:     time.Now(),
	}

	if err := repo.Append(context.Background(), c); err != nil {
		t.Fatalf("first append: %v", err)
	}

	err = repo.Append(context.Background(), c)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
}

func TestSQLiteRepository_ListEmpty(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "empty.db")

	repo, err := sqlite.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("new sqlite repo: %v", err)
	}

	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %d items", len(items))
	}
}

func TestNewSQLiteRepository_InvalidPath(t *testing.T) {
	_, err := sqlite.NewSQLiteRepository("/nonexistent/deeply/nested/path/db.sqlite")
	if err == nil {
		t.Fatal("expected error for invalid db path, got nil")
	}
}

func TestSQLiteRepository_DecimalPrecision(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "precision.db")

	repo, err := sqlite.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("new sqlite repo: %v", err)
	}

	c := domain.Caucion{
		ID:            "prec-1",
		Broker:        "Test",
		TradeDate:     time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		MaturityDate:  time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		TermDays:      1,
		Principal:     De("123456789.99"),
		TNA:           De("97.25"),
		GrossInterest: De("33356.16"),
		Fees:          De("100.50"),
		Taxes:         De("5587.83"),
		NetInterest:   De("27667.83"),
		Notes:         "precision test",
		CreatedAt:     time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
	}

	if err := repo.Append(context.Background(), c); err != nil {
		t.Fatalf("append: %v", err)
	}

	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	got := items[0]

	if got.Principal.StringFixed(2) != "123456789.99" {
		t.Fatalf("principal: got %s, want 123456789.99", got.Principal.StringFixed(2))
	}
	if got.TNA.String() != "97.25" {
		t.Fatalf("tna: got %s, want 97.25", got.TNA.String())
	}

	if !got.TradeDate.Equal(c.TradeDate) {
		t.Fatalf("trade_date: got %s, want %s", got.TradeDate, c.TradeDate)
	}
	if !got.CreatedAt.Equal(c.CreatedAt) {
		t.Fatalf("created_at: got %s, want %s", got.CreatedAt, c.CreatedAt)
	}
}

func TestSQLiteRepository_Close(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "close.db")

	repo, err := sqlite.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("new sqlite repo: %v", err)
	}

	if err := repo.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func TestSQLiteRepository_ParseDateError(t *testing.T) {
	inner := errors.New("bad date")
	e := &domain.ErrParse{Field: "trade_date", Value: "not-a-date", Err: inner}
	if e.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
	if !errors.Is(e, inner) {
		t.Fatal("expected to unwrap inner error")
	}
}
