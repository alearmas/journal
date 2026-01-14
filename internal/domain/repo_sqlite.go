package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

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

func (r *SQLiteRepository) Append(ctx context.Context, c Caucion) error {
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

func (r *SQLiteRepository) List(ctx context.Context) ([]Caucion, error) {
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
	defer rows.Close()

	var out []Caucion
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

		tradeDate, _ := time.Parse(time.RFC3339, tradeDateS)
		maturity, _ := time.Parse(time.RFC3339, maturityS)
		createdAt, _ := time.Parse(time.RFC3339, createdS)

		principal, _ := decimalFromStringSafe(principalS)
		tna, _ := decimalFromStringSafe(tnaS)
		gross, _ := decimalFromStringSafe(grossS)
		fees, _ := decimalFromStringSafe(feesS)
		taxes, _ := decimalFromStringSafe(taxesS)
		net, _ := decimalFromStringSafe(netS)

		out = append(out, Caucion{
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

// Keep it local to domain to avoid circular deps.
func decimalFromStringSafe(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}
