package service

import (
	"alearmas/tradingJournal/internal/adapter/ppi"
	"alearmas/tradingJournal/internal/domain"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// depositKeywords identify a capital deposit in PPI movement descriptions.
// Adjust after seeing real PPI descriptions.
var depositKeywords = []string{
	"acreditación",
	"acreditacion",
	"transferencia recibida",
	"depósito",
	"deposito",
	"crédito externo",
	"credito externo",
	"ingreso",
}

// withdrawalKeywords identify a capital withdrawal in PPI movement descriptions.
var withdrawalKeywords = []string{
	"débito",
	"debito",
	"retiro",
	"transferencia emitida",
	"extracción",
	"extraccion",
	"egreso",
}

// SyncResult summarises the outcome of a full sync operation.
type SyncResult struct {
	MovimientosImported int      `json:"movimientos_imported"`
	MovimientosSkipped  int      `json:"movimientos_skipped"`
	CaucionesImported   int      `json:"cauciones_imported"`
	CaucionesSkipped    int      `json:"cauciones_skipped"`
	Errors              []string `json:"errors,omitempty"`
}

// SyncService fetches data from PPI and persists it into the local journal.
type SyncService struct {
	ppiClient     *ppi.Client
	accountNumber string
	movimientoSvc *MovimientoService
	caucionSvc    *CaucionService
}

// NewSyncService creates a new SyncService.
func NewSyncService(
	client *ppi.Client,
	accountNumber string,
	movimientoSvc *MovimientoService,
	caucionSvc *CaucionService,
) *SyncService {
	return &SyncService{
		ppiClient:     client,
		accountNumber: accountNumber,
		movimientoSvc: movimientoSvc,
		caucionSvc:    caucionSvc,
	}
}

// Sync fetches both capital movements and cauciones from PPI in [dateFrom, dateTo]
// and imports records that don't already exist locally.
func (s *SyncService) Sync(ctx context.Context, dateFrom, dateTo time.Time) (SyncResult, error) {
	var result SyncResult

	// Build dedup sets upfront so we don't query the DB on every record.
	movimientosSeen, err := s.buildMovimientosSeen(ctx)
	if err != nil {
		return result, fmt.Errorf("build movimientos index: %w", err)
	}

	caucionesSeen, err := s.buildCaucionesSeen(ctx)
	if err != nil {
		return result, fmt.Errorf("build cauciones index: %w", err)
	}

	if err := s.syncMovements(ctx, dateFrom, dateTo, movimientosSeen, &result); err != nil {
		return result, fmt.Errorf("sync movements: %w", err)
	}

	if err := s.syncCauciones(ctx, dateFrom, dateTo, caucionesSeen, &result); err != nil {
		return result, fmt.Errorf("sync cauciones: %w", err)
	}

	return result, nil
}

// buildMovimientosSeen loads all PPI movimientos from the local store and returns
// a set of dedup keys: "YYYY-MM-DD|tipo|amount".
func (s *SyncService) buildMovimientosSeen(ctx context.Context) (map[string]bool, error) {
	all, err := s.movimientoSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(all))
	for _, m := range all {
		if m.Broker == "PPI" {
			seen[movimientoKey(m.Date, m.Type, m.Amount)] = true
		}
	}
	return seen, nil
}

// buildCaucionesSeen loads all PPI cauciones from the local store and returns
// a set of dedup keys: "YYYY-MM-DD|termDays|principal".
func (s *SyncService) buildCaucionesSeen(ctx context.Context) (map[string]bool, error) {
	all, err := s.caucionSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(all))
	for _, c := range all {
		if c.Broker == "PPI" {
			seen[caucionKey(c.TradeDate, c.TermDays, c.Principal)] = true
		}
	}
	return seen, nil
}

func movimientoKey(date time.Time, tipo domain.MovimientoTipo, amount decimal.Decimal) string {
	return date.Format("2006-01-02") + "|" + string(tipo) + "|" + amount.String()
}

func caucionKey(tradeDate time.Time, termDays int, principal decimal.Decimal) string {
	return tradeDate.Format("2006-01-02") + "|" + strconv.Itoa(termDays) + "|" + principal.String()
}

func (s *SyncService) syncMovements(
	ctx context.Context,
	dateFrom, dateTo time.Time,
	seen map[string]bool,
	result *SyncResult,
) error {
	movements, err := s.ppiClient.GetMovements(
		ctx, s.accountNumber,
		dateFrom.Format(time.RFC3339),
		dateTo.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	for _, m := range movements {
		tipo, ok := classifyMovimiento(m.Description)
		if !ok {
			result.MovimientosSkipped++
			continue
		}

		amount := decimal.NewFromFloat(m.Amount)
		if amount.IsNegative() {
			amount = amount.Neg()
		}

		key := movimientoKey(m.AgreementDate, tipo, amount)
		if seen[key] {
			result.MovimientosSkipped++
			continue
		}

		_, err := s.movimientoSvc.Register(ctx, RegisterMovimientoInput{
			Broker: "PPI",
			Date:   m.AgreementDate,
			Type:   tipo,
			Amount: amount,
			Notes:  m.Description,
		})
		if err != nil {
			result.Errors = append(result.Errors, "movimiento: "+err.Error())
			continue
		}

		seen[key] = true // prevent double-import within the same sync run
		result.MovimientosImported++
	}

	return nil
}

func (s *SyncService) syncCauciones(
	ctx context.Context,
	dateFrom, dateTo time.Time,
	seen map[string]bool,
	result *SyncResult,
) error {
	orders, err := s.ppiClient.GetOrders(
		ctx, s.accountNumber,
		dateFrom.Format(time.RFC3339),
		dateTo.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	for _, o := range orders {
		if !o.IsCaucion() || !o.IsSettled() {
			result.CaucionesSkipped++
			continue
		}

		// Field mapping confirmed with real PPI data:
		//   o.Amount  → Principal
		//   o.Price   → TNA (annual %, e.g. 32.1)
		//   o.Date    → TradeDate
		//   o.Ticker  → Term days: "PESOS1" → 1d, "PESOS3" → 3d, "PESOS5" → 5d
		//   NOTE: o.OperationMaxDate equals the placement date, NOT maturity.
		tradeDate := o.Date

		// Parse term days from ticker: "PESOS1" → 1, "PESOS3" → 3, "PESOS5" → 5.
		termDays := 0
		if strings.HasPrefix(o.Ticker, "PESOS") {
			if n, err := strconv.Atoi(strings.TrimPrefix(o.Ticker, "PESOS")); err == nil {
				termDays = n
			}
		}

		principal := decimal.NewFromFloat(o.Amount)
		if principal.IsNegative() {
			principal = principal.Neg()
		}
		tna := decimal.NewFromFloat(o.Price)

		if termDays <= 0 || principal.IsZero() || tna.IsZero() {
			result.CaucionesSkipped++
			result.Errors = append(result.Errors,
				fmt.Sprintf("caucion order %d skipped: missing term/principal/tna", o.ID))
			continue
		}

		key := caucionKey(tradeDate, termDays, principal)
		if seen[key] {
			result.CaucionesSkipped++
			continue
		}

		notes := fmt.Sprintf("PPI order %d | %s | %s", o.ID, o.Ticker, o.Settlement)

		_, err := s.caucionSvc.Create(ctx, CreateCaucionInput{
			Broker:    "PPI",
			TradeDate: tradeDate,
			TermDays:  termDays,
			Principal: principal,
			TNA:       tna,
			Fees:      decimal.Zero,
			Taxes:     decimal.Zero,
			Notes:     notes,
		})
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("caucion order %d: %v", o.ID, err))
			continue
		}

		seen[key] = true
		result.CaucionesImported++
	}

	return nil
}

// classifyMovimiento returns the MovimientoTipo for a PPI movement description.
func classifyMovimiento(description string) (domain.MovimientoTipo, bool) {
	lower := strings.ToLower(description)

	for _, kw := range depositKeywords {
		if strings.Contains(lower, kw) {
			return domain.Deposito, true
		}
	}
	for _, kw := range withdrawalKeywords {
		if strings.Contains(lower, kw) {
			return domain.Retiro, true
		}
	}

	return "", false
}
