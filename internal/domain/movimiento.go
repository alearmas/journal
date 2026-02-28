package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type MovimientoTipo string

const (
	Deposito MovimientoTipo = "deposito"
	Retiro   MovimientoTipo = "retiro"
)

// Movimiento representa un depósito o retiro de capital en una cuenta de broker.
type Movimiento struct {
	ID        string          `json:"id"`
	Broker    string          `json:"broker"`
	Date      time.Time       `json:"date"`
	Type      MovimientoTipo  `json:"type"`
	Amount    decimal.Decimal `json:"amount"` // siempre positivo
	Notes     string          `json:"notes"`
	CreatedAt time.Time       `json:"created_at"`
}
