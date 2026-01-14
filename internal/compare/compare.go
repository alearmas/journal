package compare

import "github.com/shopspring/decimal"

var (
	hundred    = decimal.NewFromInt(100)
	daysInYear = decimal.NewFromInt(360)
)

type CompareInput struct {
	Principal decimal.Decimal
	Days      int

	// Caucion
	CaucionTNA   decimal.Decimal
	CaucionFees  decimal.Decimal
	CaucionTaxes decimal.Decimal

	// Alternatives
	PFTNA decimal.Decimal
	MMTNA decimal.Decimal
}

type CompareOutput struct {
	CaucionNet decimal.Decimal
	PFGross    decimal.Decimal
	MMGross    decimal.Decimal
}

func SimpleGross(principal, tna decimal.Decimal, days int) decimal.Decimal {
	d := decimal.NewFromInt(int64(days))
	return principal.
		Mul(tna).
		Div(hundred).
		Mul(d).
		Div(daysInYear).
		Round(2)
}

func Run(in CompareInput) CompareOutput {
	caucionGross := SimpleGross(in.Principal, in.CaucionTNA, in.Days)
	caucionNet := caucionGross.Sub(in.CaucionFees).Sub(in.CaucionTaxes).Round(2)

	pf := SimpleGross(in.Principal, in.PFTNA, in.Days)
	mm := SimpleGross(in.Principal, in.MMTNA, in.Days)

	return CompareOutput{
		CaucionNet: caucionNet,
		PFGross:    pf,
		MMGross:    mm,
	}
}
