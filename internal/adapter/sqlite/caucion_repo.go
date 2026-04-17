// Package sqlite provides a SQLite-based driven adapter for the repository ports.
package sqlite

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"
)

// SQLiteRepository persists cauciones in a SQLite database.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository opens (or creates) the database at dbPath and runs the schema migration.
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	r := &SQLiteRepository{db: db}
	if err := r.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *SQLiteRepository) init() error {
	_, err := r.db.Exec(`
CREATE TABLE IF NOT EXISTS cauciones (
  id TEXT PRIMARY KEY,
  broker TEXT NOT NULL,
  trade_date TEXT NOT NULL,
  maturity_date TEXT NOT NULL,
  term_days INTEGER NOT NULL,
  principal TEXT NOT NULL,
  tna TEXT NOT NULL,
  gross_interest TEXT NOT NULL,
  fees TEXT NOT NULL,
  taxes TEXT NOT NULL,
  net_interest TEXT NOT NULL,
  notes TEXT NOT NULL,
  created_at TEXT NOT NULL
);
`)
	return err
}

func (r *SQLiteRepository) Append(ctx context.Context, c domain.Caucion) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO cauciones (
  id, broker, trade_date, maturity_date, term_days,
  principal, tna, gross_interest, fees, taxes, net_interest,
  notes, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		c.ID,
		c.Broker,
		c.TradeDate.Format(time.RFC3339),
		c.MaturityDate.Format(time.RFC3339),
		c.TermDays,
		c.Principal.String(),
		c.TNA.String(),
		c.GrossInterest.String(),
		c.Fees.String(),
		c.Taxes.String(),
		c.NetInterest.String(),
		c.Notes,
		c.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *SQLiteRepository) List(ctx context.Context) ([]domain.Caucion, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, broker, trade_date, maturity_date, term_days,
       principal, tna, gross_interest, fees, taxes, net_interest,
       notes, created_at
FROM cauciones
ORDER BY trade_date ASC, created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.Caucion
	for rows.Next() {
		var (
			id, broker, tradeDateS, maturityS string
			termDays                          int
			principalS, tnaS, grossS          string
			feesS, taxesS, netS               string
			notes, createdS                   string
		)

		if err := rows.Scan(
			&id, &broker, &tradeDateS, &maturityS, &termDays,
			&principalS, &tnaS, &grossS, &feesS, &taxesS, &netS,
			&notes, &createdS,
		); err != nil {
			return nil, err
		}

		tradeDate, err := time.Parse(time.RFC3339, tradeDateS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "trade_date", Value: tradeDateS, Err: err}
		}
		maturity, err := time.Parse(time.RFC3339, maturityS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "maturity_date", Value: maturityS, Err: err}
		}
		createdAt, err := time.Parse(time.RFC3339, createdS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "created_at", Value: createdS, Err: err}
		}

		principal, err := decimal.NewFromString(principalS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "principal", Value: principalS, Err: err}
		}
		tna, err := decimal.NewFromString(tnaS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "tna", Value: tnaS, Err: err}
		}
		gross, err := decimal.NewFromString(grossS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "gross_interest", Value: grossS, Err: err}
		}
		fees, err := decimal.NewFromString(feesS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "fees", Value: feesS, Err: err}
		}
		taxes, err := decimal.NewFromString(taxesS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "taxes", Value: taxesS, Err: err}
		}
		net, err := decimal.NewFromString(netS)
		if err != nil {
			return nil, &domain.ErrParse{Field: "net_interest", Value: netS, Err: err}
		}

		out = append(out, domain.Caucion{
			ID:            id,
			Broker:        broker,
			TradeDate:     tradeDate,
			MaturityDate:  maturity,
			TermDays:      termDays,
			Principal:     principal,
			TNA:           tna,
			GrossInterest: gross,
			Fees:          fees,
			Taxes:         taxes,
			NetInterest:   net,
			Notes:         notes,
			CreatedAt:     createdAt,
		})
	}
	return out, rows.Err()
}

// Close releases the underlying database connection.
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
