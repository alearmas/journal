package domain_test

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestMovimientoFileJSON_ListEmpty(t *testing.T) {
	tmp := t.TempDir()
	repo := domain.NewMovimientoFileJSONRepository(filepath.Join(tmp, "movimientos.json"))

	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty, got %d items", len(items))
	}
}

func TestMovimientoFileJSON_AppendAndList(t *testing.T) {
	tmp := t.TempDir()
	repo := domain.NewMovimientoFileJSONRepository(filepath.Join(tmp, "movimientos.json"))
	ctx := context.Background()

	m := domain.Movimiento{
		ID:        "test-id-1",
		Broker:    "Balanz",
		Date:      time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		Type:      domain.Deposito,
		Amount:    De("1000000"),
		Notes:     "test deposit",
		CreatedAt: time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC),
	}

	if err := repo.Append(ctx, m); err != nil {
		t.Fatalf("append: %v", err)
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	got := items[0]
	if got.ID != m.ID {
		t.Errorf("id: got %s, want %s", got.ID, m.ID)
	}
	if got.Broker != m.Broker {
		t.Errorf("broker: got %s, want %s", got.Broker, m.Broker)
	}
	if got.Type != m.Type {
		t.Errorf("type: got %s, want %s", got.Type, m.Type)
	}
	if !got.Amount.Equal(m.Amount) {
		t.Errorf("amount: got %s, want %s", got.Amount.String(), m.Amount.String())
	}
}

func TestMovimientoFileJSON_MultipleAppends(t *testing.T) {
	tmp := t.TempDir()
	repo := domain.NewMovimientoFileJSONRepository(filepath.Join(tmp, "movimientos.json"))
	ctx := context.Background()

	base := domain.Movimiento{
		Broker:    "Balanz",
		Date:      time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		Type:      domain.Deposito,
		Amount:    De("500000"),
		CreatedAt: time.Now(),
	}

	for i, id := range []string{"id-1", "id-2", "id-3"} {
		m := base
		m.ID = id
		_ = i
		if err := repo.Append(ctx, m); err != nil {
			t.Fatalf("append %s: %v", id, err)
		}
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3, got %d", len(items))
	}
}

func TestMovimientoFileJSON_RetiroType(t *testing.T) {
	tmp := t.TempDir()
	repo := domain.NewMovimientoFileJSONRepository(filepath.Join(tmp, "movimientos.json"))
	ctx := context.Background()

	m := domain.Movimiento{
		ID:        "retiro-1",
		Broker:    "Balanz",
		Date:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		Type:      domain.Retiro,
		Amount:    De("200000"),
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
		t.Errorf("expected Retiro type, got %s", items[0].Type)
	}
}
