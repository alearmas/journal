// Package capital computes per-broker capital summaries by crossing
// Movimiento (deposits/withdrawals) and Caucion data.
package capital

import (
	"alearmas/tradingJournal/internal/domain"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// BrokerSummary holds the capital and P&L breakdown for a single broker account.
type BrokerSummary struct {
	Broker string `json:"broker"`

	// Capital movements
	TotalDeposited decimal.Decimal `json:"total_deposited"` // sum of all deposits
	TotalWithdrawn decimal.Decimal `json:"total_withdrawn"` // sum of all withdrawals
	NetBalance     decimal.Decimal `json:"net_balance"`     // TotalDeposited - TotalWithdrawn

	// Deployment (relative to now)
	DeployedPrincipal decimal.Decimal `json:"deployed_principal"` // principal of cauciones where MaturityDate > now
	Available         decimal.Decimal `json:"available"`          // NetBalance - DeployedPrincipal

	// Returns
	TotalNetInterest decimal.Decimal `json:"total_net_interest"` // sum of NetInterest across ALL cauciones (active + matured)
	PnLPercent       decimal.Decimal `json:"pnl_percent"`        // TotalNetInterest / TotalDeposited * 100; 0 if no deposits
}

// Summarize builds a BrokerSummary per broker found in either movimientos or cauciones.
// now is used to determine whether a caución is still active (MaturityDate > now).
// Results are sorted alphabetically by broker name.
func Summarize(movimientos []domain.Movimiento, cauciones []domain.Caucion, now time.Time) []BrokerSummary {
	type acc struct {
		deposited   decimal.Decimal
		withdrawn   decimal.Decimal
		deployed    decimal.Decimal
		netInterest decimal.Decimal
	}

	data := make(map[string]*acc)

	ensure := func(broker string) {
		if _, ok := data[broker]; !ok {
			data[broker] = &acc{}
		}
	}

	for _, m := range movimientos {
		ensure(m.Broker)
		switch m.Type {
		case domain.Deposito:
			data[m.Broker].deposited = data[m.Broker].deposited.Add(m.Amount)
		case domain.Retiro:
			data[m.Broker].withdrawn = data[m.Broker].withdrawn.Add(m.Amount)
		}
	}

	for _, c := range cauciones {
		ensure(c.Broker)
		// Only count as deployed if the caución has not yet matured.
		if c.MaturityDate.After(now) {
			data[c.Broker].deployed = data[c.Broker].deployed.Add(c.Principal)
		}
		data[c.Broker].netInterest = data[c.Broker].netInterest.Add(c.NetInterest)
	}

	hundred := decimal.NewFromInt(100)
	summaries := make([]BrokerSummary, 0, len(data))

	for broker, a := range data {
		net := a.deposited.Sub(a.withdrawn)
		available := net.Sub(a.deployed)

		var pnl decimal.Decimal
		if a.deposited.IsPositive() {
			pnl = a.netInterest.Div(a.deposited).Mul(hundred).Round(4)
		}

		summaries = append(summaries, BrokerSummary{
			Broker:            broker,
			TotalDeposited:    a.deposited.Round(2),
			TotalWithdrawn:    a.withdrawn.Round(2),
			NetBalance:        net.Round(2),
			DeployedPrincipal: a.deployed.Round(2),
			Available:         available.Round(2),
			TotalNetInterest:  a.netInterest.Round(2),
			PnLPercent:        pnl,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Broker < summaries[j].Broker
	})

	return summaries
}
