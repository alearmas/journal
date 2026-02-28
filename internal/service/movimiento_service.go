package service

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type RegisterMovimientoInput struct {
	Broker string
	Date   time.Time
	Type   domain.MovimientoTipo
	Amount decimal.Decimal
	Notes  string
}

type MovimientoService struct {
	repo domain.MovimientoRepository
}

func NewMovimientoService(repo domain.MovimientoRepository) *MovimientoService {
	return &MovimientoService{repo: repo}
}

func (s *MovimientoService) Register(ctx context.Context, in RegisterMovimientoInput) (domain.Movimiento, error) {
	if in.Amount.Cmp(decimal.Zero) <= 0 {
		return domain.Movimiento{}, &domain.ErrValidation{Field: "amount", Message: "must be > 0"}
	}
	if in.Type != domain.Deposito && in.Type != domain.Retiro {
		return domain.Movimiento{}, &domain.ErrValidation{Field: "type", Message: "must be deposito or retiro"}
	}
	if in.Broker == "" {
		in.Broker = "Balanz"
	}

	date := in.Date
	if date.IsZero() {
		date = time.Now()
	}

	id, err := newID()
	if err != nil {
		return domain.Movimiento{}, err
	}

	m := domain.Movimiento{
		ID:        id,
		Broker:    in.Broker,
		Date:      date,
		Type:      in.Type,
		Amount:    in.Amount.Round(2),
		Notes:     in.Notes,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Append(ctx, m); err != nil {
		return domain.Movimiento{}, err
	}
	return m, nil
}

func (s *MovimientoService) List(ctx context.Context) ([]domain.Movimiento, error) {
	return s.repo.List(ctx)
}
