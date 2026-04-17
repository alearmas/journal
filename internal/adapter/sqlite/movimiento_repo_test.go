package sqlite_test

import (
	"alearmas/tradingJournal/internal/adapter/sqlite"
	"alearmas/tradingJournal/internal/domain"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestMovimientoSQLite_ListEmpty(t *testing.T) {
	tmp := t.TempDir()
	repo, err := sqlite.NewMovimientoSQLiteRepository(filepath.Join(tmp, "mov.db"))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	defer func() { _ = repo.Close() }()

	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty, got %d", len(items))
	}
}

func TestMovimientoSQLite_AppendAndList(t *testing.T) {
	tmp := t.TempDir()
	repo, err := sqlite.NewMovimientoSQLiteRepository(filepath.Join(tmp, "mov.db"))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	m := domain.Movimiento{
		ID:        "sqlite-dep-1",
		Broker:    "Balanz",
		Date:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:      domain.Deposito,
		Amount:    De("2000000"),
		Notes:     "sqlite test",
		CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	if err := repo.Append(ctx, m); err != nil {
		t.Fatalf("append: %v", err)
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1, got %d", len(items))
	}

	got := items[0]
	if got.ID != m.ID {
		t.Errorf("id: got %s, want %s", got.ID, m.ID)
	}
	if got.Type != domain.Deposito {
		t.Errorf("type: got %s, want deposito", got.Type)
	}
	if got.Amount.StringFixed(2) != "2000000.00" {
		t.Errorf("amount: got %s, want 2000000.00", got.Amount.StringFixed(2))
	}
	if got.Notes != m.Notes {
		t.Errorf("notes: got %s, want %s", got.Notes, m.Notes)
	}
	if !got.Date.Equal(m.Date) {
		t.Errorf("date: got %s, want %s", got.Date, m.Date)
	}
}

func TestMovimientoSQLite_RetiroType(t *testing.T) {
	tmp := t.TempDir()
	repo, err := sqlite.NewMovimientoSQLiteRepository(filepath.Join(tmp, "mov.db"))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	m := domain.Movimiento{
		ID:        "retiro-sqlite-1",
		Broker:    "Balanz",
		Date:      time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		Type:      domain.Retiro,
		Amount:    De("300000"),
		CreatedAt: time.Now(),
	}

	if err := repo.Append(ctx, m); err != nil {
		t.Fatalf("append: %v", err)
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if items[0].Type != domain.Retiro {
		t.Errorf("expected Retiro, got %s", items[0].Type)
	}
}

func TestMovimientoSQLite_DuplicateIDError(t *testing.T) {
	tmp := t.TempDir()
	repo, err := sqlite.NewMovimientoSQLiteRepository(filepath.Join(tmp, "dup.db"))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	m := domain.Movimiento{
		ID:        "dup-mov-1",
		Broker:    "Balanz",
		Date:      time.Now(),
		Type:      domain.Deposito,
		Amount:    De("100000"),
		CreatedAt: time.Now(),
	}

	if err := repo.Append(ctx, m); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := repo.Append(ctx, m); err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
}

func TestMovimientoSQLite_Close(t *testing.T) {
	tmp := t.TempDir()
	repo, err := sqlite.NewMovimientoSQLiteRepository(filepath.Join(tmp, "close.db"))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func TestMovimientoSQLite_InvalidPath(t *testing.T) {
	_, err := sqlite.NewMovimientoSQLiteRepository("/nonexistent/deeply/nested/path/mov.sqlite")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}
