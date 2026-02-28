package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"
)

type MovimientoSQLiteRepository struct {
	db *sql.DB
}

func NewMovimientoSQLiteRepository(dbPath string) (*MovimientoSQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	r := &MovimientoSQLiteRepository{db: db}
	if err := r.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *MovimientoSQLiteRepository) init() error {
	_, err := r.db.Exec(`
CREATE TABLE IF NOT EXISTS movimientos (
  id TEXT PRIMARY KEY,
  broker TEXT NOT NULL,
  date TEXT NOT NULL,
  type TEXT NOT NULL,
  amount TEXT NOT NULL,
  notes TEXT NOT NULL,
  created_at TEXT NOT NULL
);
`)
	return err
}

func (r *MovimientoSQLiteRepository) Append(ctx context.Context, m Movimiento) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO movimientos (id, broker, date, type, amount, notes, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
`,
		m.ID,
		m.Broker,
		m.Date.Format(time.RFC3339),
		string(m.Type),
		m.Amount.String(),
		m.Notes,
		m.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *MovimientoSQLiteRepository) List(ctx context.Context) ([]Movimiento, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, broker, date, type, amount, notes, created_at
FROM movimientos
ORDER BY date ASC, created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Movimiento
	for rows.Next() {
		var (
			id, broker, dateS, typ string
			amountS, notes, createdS string
		)

		if err := rows.Scan(&id, &broker, &dateS, &typ, &amountS, &notes, &createdS); err != nil {
			return nil, err
		}

		date, err := time.Parse(time.RFC3339, dateS)
		if err != nil {
			return nil, &ErrParse{Field: "date", Value: dateS, Err: err}
		}
		createdAt, err := time.Parse(time.RFC3339, createdS)
		if err != nil {
			return nil, &ErrParse{Field: "created_at", Value: createdS, Err: err}
		}
		amount, err := decimal.NewFromString(amountS)
		if err != nil {
			return nil, &ErrParse{Field: "amount", Value: amountS, Err: err}
		}

		out = append(out, Movimiento{
			ID:        id,
			Broker:    broker,
			Date:      date,
			Type:      MovimientoTipo(typ),
			Amount:    amount,
			Notes:     notes,
			CreatedAt: createdAt,
		})
	}
	return out, rows.Err()
}

func (r *MovimientoSQLiteRepository) Close() error {
	return r.db.Close()
}
