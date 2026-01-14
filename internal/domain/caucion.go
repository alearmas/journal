package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Caucion struct {
	ID            string          `json:"id"`
	Broker        string          `json:"broker"`
	TradeDate     time.Time       `json:"trade_date"`
	MaturityDate  time.Time       `json:"maturity_date"`
	TermDays      int             `json:"term_days"`
	Principal     decimal.Decimal `json:"principal"`
	TNA           decimal.Decimal `json:"tna"` // Annual nominal rate, percentage (e.g., 85.5)
	GrossInterest decimal.Decimal `json:"gross_interest"`
	Fees          decimal.Decimal `json:"fees"`
	Taxes         decimal.Decimal `json:"taxes"`
	NetInterest   decimal.Decimal `json:"net_interest"`
	Notes         string          `json:"notes"`
	CreatedAt     time.Time       `json:"created_at"`
}
