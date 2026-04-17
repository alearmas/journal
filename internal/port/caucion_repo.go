// Package port defines the secondary ports (output) of the application.
// These interfaces are implemented by driven adapters (filejson, sqlite).
package port

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
)

// CaucionRepository is the secondary port for persisting and retrieving cauciones.
type CaucionRepository interface {
	List(ctx context.Context) ([]domain.Caucion, error)
	Append(ctx context.Context, c domain.Caucion) error
}
