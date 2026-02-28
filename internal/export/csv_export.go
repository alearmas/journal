package export

import (
	"alearmas/tradingJournal/internal/domain"
	"encoding/csv"
	"os"
	"strconv"
)

func WriteCaucionesCSV(path string, items []domain.Caucion) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"id", "broker", "trade_date", "maturity_date", "term_days",
		"principal", "tna", "gross_interest", "fees", "taxes", "net_interest",
		"notes", "created_at",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, c := range items {
		row := []string{
			c.ID,
			c.Broker,
			c.TradeDate.Format("2006-01-02"),
			c.MaturityDate.Format("2006-01-02"),
			intToString(c.TermDays),
			c.Principal.StringFixed(2),
			c.TNA.String(),
			c.GrossInterest.StringFixed(2),
			c.Fees.StringFixed(2),
			c.Taxes.StringFixed(2),
			c.NetInterest.StringFixed(2),
			c.Notes,
			c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return w.Error()
}

func intToString(n int) string {
	return strconv.Itoa(n)
}
