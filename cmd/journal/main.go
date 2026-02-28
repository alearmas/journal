package main

import (
	"alearmas/tradingJournal/internal/compare"
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/export"
	"alearmas/tradingJournal/internal/report"
	"alearmas/tradingJournal/internal/service"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

func die(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(code)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	dataPath := getenv("JOURNAL_DATA", "data/cauciones.json")
	store := getenv("JOURNAL_STORE", "json")

	var repo domain.CaucionRepository
	switch store {
	case "json":
		repo = domain.NewFileJSONRepository(dataPath)
	case "sqlite":
		dbPath := getenv("JOURNAL_DB", "data/journal.db")
		sqlRepo, err := domain.NewSQLiteRepository(dbPath)
		if err != nil {
			die(1, "error opening sqlite: %v", err)
		}
		repo = sqlRepo
	default:
		die(2, "invalid JOURNAL_STORE, use json|sqlite")
	}

	if closer, ok := repo.(io.Closer); ok {
		defer closer.Close()
	}

	svc := service.NewCaucionService(repo)

	switch os.Args[1] {
	case "add":
		addCmd(svc, os.Args[2:])
	case "list":
		listCmd(svc)
	case "summary":
		summaryCmd(svc)
	case "report":
		reportCmd(svc, os.Args[2:])
	case "export":
		exportCmd(svc, os.Args[2:])
	case "compare":
		compareCmd(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`Usage:
  journal add --principal 1000000.00 --tna 85.5 --term 1 --fees 50.00 --taxes 421.00 --date 2026-01-10 --notes "overnight"
  journal list
  journal summary

Notes:
  - Monetary values should be passed as strings compatible with decimal (e.g. 1000000, 1000000.00)
  - TNA is a percentage (e.g. 85.5)

Env:
  JOURNAL_DATA=path/to/cauciones.json (default: data/cauciones.json)
`)
}

func addCmd(svc *service.CaucionService, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)

	var (
		broker = fs.String("broker", "Balanz", "Broker name")

		// Use strings for decimal-safe parsing (no float64 anywhere).
		principalStr = fs.String("principal", "0", "Principal amount (decimal), e.g. 1000000.00")
		tnaStr       = fs.String("tna", "0", "Annual nominal rate (TNA %) (decimal), e.g. 85.5")
		feesStr      = fs.String("fees", "0", "Fees/commissions (decimal)")
		taxesStr     = fs.String("taxes", "0", "Taxes/withholdings (decimal)")

		term    = fs.Int("term", 0, "Term in days")
		dateStr = fs.String("date", "", "Trade date YYYY-MM-DD (optional)")
		notes   = fs.String("notes", "", "Notes")
	)

	_ = fs.Parse(args)

	principal, err := decimal.NewFromString(*principalStr)
	if err != nil {
		die(2, "invalid --principal, expected decimal (e.g. 1000000.00)")
	}
	tna, err := decimal.NewFromString(*tnaStr)
	if err != nil {
		die(2, "invalid --tna, expected decimal (e.g. 85.5)")
	}
	fees, err := decimal.NewFromString(*feesStr)
	if err != nil {
		die(2, "invalid --fees, expected decimal (e.g. 50.00)")
	}
	taxes, err := decimal.NewFromString(*taxesStr)
	if err != nil {
		die(2, "invalid --taxes, expected decimal (e.g. 421.00)")
	}

	var tradeDate time.Time
	if *dateStr != "" {
		d, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			die(2, "invalid --date, expected YYYY-MM-DD")
		}
		tradeDate = d
	}

	c, err := svc.Create(context.Background(), service.CreateCaucionInput{
		Broker:    *broker,
		TradeDate: tradeDate,
		TermDays:  *term,
		Principal: principal,
		TNA:       tna,
		Fees:      fees,
		Taxes:     taxes,
		Notes:     *notes,
	})
	if err != nil {
		die(1, "error: %v", err)
	}

	fmt.Printf("Saved caucion %s\n", c.ID)
	fmt.Printf("Trade: %s  Maturity: %s  TermDays: %d\n",
		c.TradeDate.Format("2006-01-02"),
		c.MaturityDate.Format("2006-01-02"),
		c.TermDays,
	)

	fmt.Printf("Principal: %s  TNA: %s%%\n",
		c.Principal.StringFixed(2),
		c.TNA.String(),
	)

	fmt.Printf("GrossInterest: %s  Fees: %s  Taxes: %s  NetInterest: %s\n",
		c.GrossInterest.StringFixed(2),
		c.Fees.StringFixed(2),
		c.Taxes.StringFixed(2),
		c.NetInterest.StringFixed(2),
	)
}

func listCmd(svc *service.CaucionService) {
	items, err := svc.List(context.Background())
	if err != nil {
		die(1, "error: %v", err)
	}
	if len(items) == 0 {
		fmt.Println("No cauciones found.")
		return
	}

	for _, c := range items {
		fmt.Printf(
			"%s | %s -> %s | term=%dd | principal=%s | tna=%s%% | net=%s | notes=%s\n",
			c.ID,
			c.TradeDate.Format("2006-01-02"),
			c.MaturityDate.Format("2006-01-02"),
			c.TermDays,
			c.Principal.StringFixed(2),
			c.TNA.String(),
			c.NetInterest.StringFixed(2),
			c.Notes,
		)
	}
}

func summaryCmd(svc *service.CaucionService) {
	items, err := svc.List(context.Background())
	if err != nil {
		die(1, "error: %v", err)
	}

	totalPrincipal := decimal.Zero
	totalNet := decimal.Zero
	totalGross := decimal.Zero
	totalFees := decimal.Zero
	totalTaxes := decimal.Zero

	for _, c := range items {
		totalPrincipal = totalPrincipal.Add(c.Principal)
		totalGross = totalGross.Add(c.GrossInterest)
		totalFees = totalFees.Add(c.Fees)
		totalTaxes = totalTaxes.Add(c.Taxes)
		totalNet = totalNet.Add(c.NetInterest)
	}

	fmt.Printf("Count: %d\n", len(items))
	fmt.Printf("Total Principal: %s\n", totalPrincipal.StringFixed(2))
	fmt.Printf("Total Gross Interest: %s\n", totalGross.StringFixed(2))
	fmt.Printf("Total Fees: %s\n", totalFees.StringFixed(2))
	fmt.Printf("Total Taxes: %s\n", totalTaxes.StringFixed(2))
	fmt.Printf("Total Net Interest: %s\n", totalNet.StringFixed(2))
}

func reportCmd(svc *service.CaucionService, args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	month := fs.String("month", "", "Month YYYY-MM (required)")
	_ = fs.Parse(args)

	if *month == "" {
		die(2, "missing --month YYYY-MM")
	}

	items, err := svc.List(context.Background())
	if err != nil {
		die(1, "error: %v", err)
	}

	filtered, err := report.FilterByMonth(items, *month)
	if err != nil {
		die(2, "invalid --month: %v", err)
	}

	s := report.SummarizeMonth(filtered, *month)
	fmt.Printf("Month: %s\n", s.Month)
	fmt.Printf("Count: %d\n", s.Count)
	fmt.Printf("Total Principal: %s\n", s.TotalPrincipal.StringFixed(2))
	fmt.Printf("Total Gross Interest: %s\n", s.TotalGross.StringFixed(2))
	fmt.Printf("Total Fees: %s\n", s.TotalFees.StringFixed(2))
	fmt.Printf("Total Taxes: %s\n", s.TotalTaxes.StringFixed(2))
	fmt.Printf("Total Net Interest: %s\n", s.TotalNet.StringFixed(2))
	fmt.Printf("Weighted Avg TNA: %s%%\n", s.WeightedAvgTNA.String())
}

func exportCmd(svc *service.CaucionService, args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	out := fs.String("out", "cauciones.csv", "Output CSV path")
	month := fs.String("month", "", "Optional month YYYY-MM")
	_ = fs.Parse(args)

	items, err := svc.List(context.Background())
	if err != nil {
		die(1, "error: %v", err)
	}

	if *month != "" {
		filtered, err := report.FilterByMonth(items, *month)
		if err != nil {
			die(2, "invalid --month: %v", err)
		}
		items = filtered
	}

	if err := export.WriteCaucionesCSV(*out, items); err != nil {
		die(1, "error writing csv: %v", err)
	}

	fmt.Println("CSV exported to:", *out)
}

func compareCmd(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)

	principalStr := fs.String("principal", "0", "Principal (decimal)")
	days := fs.Int("days", 1, "Days")

	cTnaStr := fs.String("caucion-tna", "0", "Caucion TNA (decimal %)")
	feesStr := fs.String("fees", "0", "Caucion fees (decimal)")
	taxesStr := fs.String("taxes", "0", "Caucion taxes (decimal)")

	pfTnaStr := fs.String("pf-tna", "0", "Plazo fijo TNA (decimal %)")
	mmTnaStr := fs.String("mm-tna", "0", "Money market TNA (decimal %)")

	_ = fs.Parse(args)

	principal, err := decimal.NewFromString(*principalStr)
	if err != nil {
		die(2, "invalid --principal")
	}
	cTna, err := decimal.NewFromString(*cTnaStr)
	if err != nil {
		die(2, "invalid --caucion-tna")
	}
	fees, err := decimal.NewFromString(*feesStr)
	if err != nil {
		die(2, "invalid --fees")
	}
	taxes, err := decimal.NewFromString(*taxesStr)
	if err != nil {
		die(2, "invalid --taxes")
	}
	pfTna, err := decimal.NewFromString(*pfTnaStr)
	if err != nil {
		die(2, "invalid --pf-tna")
	}
	mmTna, err := decimal.NewFromString(*mmTnaStr)
	if err != nil {
		die(2, "invalid --mm-tna")
	}

	out := compare.Run(compare.CompareInput{
		Principal:    principal,
		Days:         *days,
		CaucionTNA:   cTna,
		CaucionFees:  fees,
		CaucionTaxes: taxes,
		PFTNA:        pfTna,
		MMTNA:        mmTna,
	})

	fmt.Printf("Days: %d  Principal: %s\n", *days, principal.StringFixed(2))
	fmt.Printf("Caucion net: %s\n", out.CaucionNet.StringFixed(2))
	fmt.Printf("PF gross:    %s\n", out.PFGross.StringFixed(2))
	fmt.Printf("MM gross:    %s\n", out.MMGross.StringFixed(2))
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
