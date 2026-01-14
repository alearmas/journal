package report

import (
	"alearmas/tradingJournal/internal/domain"
	"time"

	"github.com/shopspring/decimal"
)

type MonthSummary struct {
	Month string

	Count int

	TotalPrincipal decimal.Decimal
	TotalGross     decimal.Decimal
	TotalFees      decimal.Decimal
	TotalTaxes     decimal.Decimal
	TotalNet       decimal.Decimal

	// Weighted average TNA by principal (simple, practical).
	WeightedAvgTNA decimal.Decimal
}

func FilterByMonth(items []domain.Caucion, month string) ([]domain.Caucion, error) {
	// month: YYYY-MM
	_, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Caucion, 0, len(items))
	for _, c := range items {
		if c.TradeDate.Format("2006-01") == month {
			out = append(out, c)
		}
	}
	return out, nil
}

func SummarizeMonth(items []domain.Caucion, month string) MonthSummary {
	totalPrincipal := decimal.Zero
	totalGross := decimal.Zero
	totalFees := decimal.Zero
	totalTaxes := decimal.Zero
	totalNet := decimal.Zero

	weightedTNASum := decimal.Zero // sum(principal * tna)
	for _, c := range items {
		totalPrincipal = totalPrincipal.Add(c.Principal)
		totalGross = totalGross.Add(c.GrossInterest)
		totalFees = totalFees.Add(c.Fees)
		totalTaxes = totalTaxes.Add(c.Taxes)
		totalNet = totalNet.Add(c.NetInterest)

		weightedTNASum = weightedTNASum.Add(c.Principal.Mul(c.TNA))
	}

	avgTNA := decimal.Zero
	if totalPrincipal.Cmp(decimal.Zero) > 0 {
		avgTNA = weightedTNASum.Div(totalPrincipal).Round(4)
	}

	return MonthSummary{
		Month:          month,
		Count:          len(items),
		TotalPrincipal: totalPrincipal.Round(2),
		TotalGross:     totalGross.Round(2),
		TotalFees:      totalFees.Round(2),
		TotalTaxes:     totalTaxes.Round(2),
		TotalNet:       totalNet.Round(2),
		WeightedAvgTNA: avgTNA,
	}
}
