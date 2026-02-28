package domain

import "context"

type MovimientoRepository interface {
	List(ctx context.Context) ([]Movimiento, error)
	Append(ctx context.Context, m Movimiento) error
}
