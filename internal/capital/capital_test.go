package capital_test

import (
	"alearmas/tradingJournal/internal/capital"
	"alearmas/tradingJournal/internal/domain"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func D(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic("test helper D: " + err.Error())
	}
	return d
}

var now = time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

func mov(broker string, typ domain.MovimientoTipo, amount string) domain.Movimiento {
	return domain.Movimiento{
		ID:        "m-" + broker + "-" + string(typ),
		Broker:    broker,
		Date:      now.AddDate(0, -1, 0),
		Type:      typ,
		Amount:    D(amount),
		CreatedAt: now.AddDate(0, -1, 0),
	}
}

func caucion(broker, principal, net string, maturity time.Time) domain.Caucion {
	return domain.Caucion{
		ID:            "c-" + broker,
		Broker:        broker,
		TradeDate:     now.AddDate(0, -1, 0),
		MaturityDate:  maturity,
		TermDays:      30,
		Principal:     D(principal),
		TNA:           D("85"),
		GrossInterest: D(net),
		Fees:          decimal.Zero,
		Taxes:         decimal.Zero,
		NetInterest:   D(net),
		CreatedAt:     now.AddDate(0, -1, 0),
	}
}

func findBroker(summaries []capital.BrokerSummary, name string) (capital.BrokerSummary, bool) {
	for _, s := range summaries {
		if s.Broker == name {
			return s, true
		}
	}
	return capital.BrokerSummary{}, false
}

func TestSummarize_Empty(t *testing.T) {
	result := capital.Summarize(nil, nil, now)
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d entries", len(result))
	}
}

func TestSummarize_OnlyDeposit_NoCauciones(t *testing.T) {
	movs := []domain.Movimiento{mov("Balanz", domain.Deposito, "1000000")}
	result := capital.Summarize(movs, nil, now)

	s, ok := findBroker(result, "Balanz")
	if !ok {
		t.Fatal("expected Balanz in result")
	}

	if s.TotalDeposited.StringFixed(2) != "1000000.00" {
		t.Errorf("deposited: got %s, want 1000000.00", s.TotalDeposited.StringFixed(2))
	}
	if !s.TotalWithdrawn.IsZero() {
		t.Errorf("withdrawn should be zero, got %s", s.TotalWithdrawn.String())
	}
	if s.NetBalance.StringFixed(2) != "1000000.00" {
		t.Errorf("net balance: got %s, want 1000000.00", s.NetBalance.StringFixed(2))
	}
	if !s.DeployedPrincipal.IsZero() {
		t.Errorf("deployed should be zero, got %s", s.DeployedPrincipal.String())
	}
	if s.Available.StringFixed(2) != "1000000.00" {
		t.Errorf("available: got %s, want 1000000.00", s.Available.StringFixed(2))
	}
	if !s.PnLPercent.IsZero() {
		t.Errorf("pnl should be zero with no cauciones, got %s", s.PnLPercent.String())
	}
}

func TestSummarize_DepositAndWithdraw(t *testing.T) {
	movs := []domain.Movimiento{
		mov("Balanz", domain.Deposito, "1000000"),
		mov("Balanz", domain.Retiro, "200000"),
	}
	result := capital.Summarize(movs, nil, now)

	s, ok := findBroker(result, "Balanz")
	if !ok {
		t.Fatal("expected Balanz in result")
	}

	if s.NetBalance.StringFixed(2) != "800000.00" {
		t.Errorf("net balance: got %s, want 800000.00", s.NetBalance.StringFixed(2))
	}
	if s.Available.StringFixed(2) != "800000.00" {
		t.Errorf("available: got %s, want 800000.00", s.Available.StringFixed(2))
	}
}

func TestSummarize_ActiveCaucionReducesAvailable(t *testing.T) {
	// Caución matures in the future → still active
	activeMaturiy := now.AddDate(0, 0, 10)
	movs := []domain.Movimiento{mov("Balanz", domain.Deposito, "1000000")}
	cauciones := []domain.Caucion{caucion("Balanz", "800000", "5000", activeMaturiy)}

	result := capital.Summarize(movs, cauciones, now)
	s, ok := findBroker(result, "Balanz")
	if !ok {
		t.Fatal("expected Balanz in result")
	}

	if s.DeployedPrincipal.StringFixed(2) != "800000.00" {
		t.Errorf("deployed: got %s, want 800000.00", s.DeployedPrincipal.StringFixed(2))
	}
	if s.Available.StringFixed(2) != "200000.00" {
		t.Errorf("available: got %s, want 200000.00", s.Available.StringFixed(2))
	}
}

func TestSummarize_MaturedCaucionDoesNotReduceAvailable(t *testing.T) {
	// Caución already matured → money is back, not deployed
	pastMaturity := now.AddDate(0, 0, -1)
	movs := []domain.Movimiento{mov("Balanz", domain.Deposito, "1000000")}
	cauciones := []domain.Caucion{caucion("Balanz", "800000", "5000", pastMaturity)}

	result := capital.Summarize(movs, cauciones, now)
	s, ok := findBroker(result, "Balanz")
	if !ok {
		t.Fatal("expected Balanz in result")
	}

	if !s.DeployedPrincipal.IsZero() {
		t.Errorf("matured caución should not count as deployed, got %s", s.DeployedPrincipal.String())
	}
	if s.Available.StringFixed(2) != "1000000.00" {
		t.Errorf("available: got %s, want 1000000.00", s.Available.StringFixed(2))
	}
}

func TestSummarize_PnLCalculation(t *testing.T) {
	// 1000000 deposited, 5000 net interest → P&L = 0.5%
	pastMaturity := now.AddDate(0, 0, -1)
	movs := []domain.Movimiento{mov("Balanz", domain.Deposito, "1000000")}
	cauciones := []domain.Caucion{caucion("Balanz", "1000000", "5000", pastMaturity)}

	result := capital.Summarize(movs, cauciones, now)
	s, ok := findBroker(result, "Balanz")
	if !ok {
		t.Fatal("expected Balanz in result")
	}

	if s.TotalNetInterest.StringFixed(2) != "5000.00" {
		t.Errorf("net interest: got %s, want 5000.00", s.TotalNetInterest.StringFixed(2))
	}
	if s.PnLPercent.StringFixed(4) != "0.5000" {
		t.Errorf("P&L: got %s, want 0.5000", s.PnLPercent.StringFixed(4))
	}
}

func TestSummarize_NoPnLWithoutDeposits(t *testing.T) {
	// Caución without any deposit recorded → P&L stays 0 (no division by zero)
	past := now.AddDate(0, 0, -1)
	cauciones := []domain.Caucion{caucion("Balanz", "1000000", "5000", past)}

	result := capital.Summarize(nil, cauciones, now)
	s, ok := findBroker(result, "Balanz")
	if !ok {
		t.Fatal("expected Balanz in result")
	}

	if !s.PnLPercent.IsZero() {
		t.Errorf("expected PnL 0 without deposits, got %s", s.PnLPercent.String())
	}
}

func TestSummarize_MultipleBrokers(t *testing.T) {
	movs := []domain.Movimiento{
		mov("Balanz", domain.Deposito, "1000000"),
		mov("BYMA", domain.Deposito, "500000"),
	}
	result := capital.Summarize(movs, nil, now)

	if len(result) != 2 {
		t.Fatalf("expected 2 brokers, got %d", len(result))
	}
	// Results should be sorted alphabetically
	if result[0].Broker != "BYMA" || result[1].Broker != "Balanz" {
		t.Errorf("expected BYMA then Balanz (alphabetical), got %s then %s",
			result[0].Broker, result[1].Broker)
	}
}
