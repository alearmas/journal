package port

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
)

// MovimientoRepository is the secondary port for persisting and retrieving movimientos.
type MovimientoRepository interface {
	List(ctx context.Context) ([]domain.Movimiento, error)
	Append(ctx context.Context, m domain.Movimiento) error
}
