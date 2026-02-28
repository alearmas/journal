package service_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/service"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// movimientoMemRepo is an in-memory MovimientoRepository for testing.
type movimientoMemRepo struct {
	items []domain.Movimiento
	err   error // if non-nil, Append returns this error
}

func (r *movimientoMemRepo) List(_ context.Context) ([]domain.Movimiento, error) {
	return r.items, nil
}

func (r *movimientoMemRepo) Append(_ context.Context, m domain.Movimiento) error {
	if r.err != nil {
		return r.err
	}
	r.items = append(r.items, m)
	return nil
}

func validMovInput(typ domain.MovimientoTipo) service.RegisterMovimientoInput {
	return service.RegisterMovimientoInput{
		Broker: "Balanz",
		Date:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		Type:   typ,
		Amount: decimal.NewFromInt(1000000),
		Notes:  "test",
	}
}

func TestMovimientoService_RegisterDeposito(t *testing.T) {
	repo := &movimientoMemRepo{}
	msvc := service.NewMovimientoService(repo)

	m, err := msvc.Register(context.Background(), validMovInput(domain.Deposito))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if m.Type != domain.Deposito {
		t.Errorf("type: got %s, want deposito", m.Type)
	}
	if m.Amount.StringFixed(2) != "1000000.00" {
		t.Errorf("amount: got %s, want 1000000.00", m.Amount.StringFixed(2))
	}
	if m.Broker != "Balanz" {
		t.Errorf("broker: got %s, want Balanz", m.Broker)
	}
}

func TestMovimientoService_RegisterRetiro(t *testing.T) {
	repo := &movimientoMemRepo{}
	msvc := service.NewMovimientoService(repo)

	m, err := msvc.Register(context.Background(), validMovInput(domain.Retiro))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Type != domain.Retiro {
		t.Errorf("type: got %s, want retiro", m.Type)
	}
}

func TestMovimientoService_ValidationErrors(t *testing.T) {
	repo := &movimientoMemRepo{}
	msvc := service.NewMovimientoService(repo)
	ctx := context.Background()

	tests := []struct {
		name  string
		mod   func(*service.RegisterMovimientoInput)
		field string
	}{
		{"zero amount", func(in *service.RegisterMovimientoInput) { in.Amount = decimal.Zero }, "amount"},
		{"negative amount", func(in *service.RegisterMovimientoInput) { in.Amount = decimal.NewFromInt(-1) }, "amount"},
		{"invalid type", func(in *service.RegisterMovimientoInput) { in.Type = "pagare" }, "type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validMovInput(domain.Deposito)
			tt.mod(&in)

			_, err := msvc.Register(ctx, in)
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

func TestMovimientoService_DefaultBroker(t *testing.T) {
	repo := &movimientoMemRepo{}
	msvc := service.NewMovimientoService(repo)

	in := validMovInput(domain.Deposito)
	in.Broker = ""

	m, err := msvc.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Broker != "Balanz" {
		t.Errorf("expected default broker Balanz, got %s", m.Broker)
	}
}

func TestMovimientoService_DefaultDate(t *testing.T) {
	repo := &movimientoMemRepo{}
	msvc := service.NewMovimientoService(repo)

	in := validMovInput(domain.Deposito)
	in.Date = time.Time{} // zero

	m, err := msvc.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Date.IsZero() {
		t.Fatal("expected non-zero date")
	}
}

func TestMovimientoService_RepoError(t *testing.T) {
	repoErr := errors.New("disk full")
	repo := &movimientoMemRepo{err: repoErr}
	msvc := service.NewMovimientoService(repo)

	_, err := msvc.Register(context.Background(), validMovInput(domain.Deposito))
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}

func TestMovimientoService_List(t *testing.T) {
	repo := &movimientoMemRepo{}
	msvc := service.NewMovimientoService(repo)
	ctx := context.Background()

	if _, err := msvc.Register(ctx, validMovInput(domain.Deposito)); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := msvc.Register(ctx, validMovInput(domain.Retiro)); err != nil {
		t.Fatalf("register: %v", err)
	}

	items, err := msvc.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
}
