package service

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type CreateCaucionInput struct {
	Broker    string
	TradeDate time.Time
	TermDays  int
	Principal decimal.Decimal
	TNA       decimal.Decimal
	Fees      decimal.Decimal
	Taxes     decimal.Decimal
	Notes     string
}

type CaucionService struct {
	repo domain.CaucionRepository
}

func NewCaucionService(repo domain.CaucionRepository) *CaucionService {
	return &CaucionService{repo: repo}
}

var (
	hundred    = decimal.NewFromInt(100)
	daysInYear = decimal.NewFromInt(360)
)

// GrossInterest = Principal * (TNA/100) * (TermDays/365)
func ComputeGrossInterest(principal, tna decimal.Decimal, termDays int) decimal.Decimal {
	td := decimal.NewFromInt(int64(termDays))

	return principal.
		Mul(tna).
		Div(hundred).
		Mul(td).
		Div(daysInYear).
		Round(2)
}

func round2(x decimal.Decimal) decimal.Decimal {
	return x.Round(2)
}

func newID() (string, error) {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *CaucionService) Create(ctx context.Context, in CreateCaucionInput) (domain.Caucion, error) {
	if in.TermDays <= 0 {
		return domain.Caucion{}, errors.New("termDays must be > 0")
	}
	if in.Principal.Cmp(decimal.Zero) <= 0 {
		return domain.Caucion{}, errors.New("principal must be > 0")
	}
	if in.TNA.Cmp(decimal.Zero) <= 0 {
		return domain.Caucion{}, errors.New("TNA must be > 0")
	}
	if in.Fees.Cmp(decimal.Zero) < 0 {
		return domain.Caucion{}, errors.New("fees must be >= 0")
	}
	if in.Taxes.Cmp(decimal.Zero) < 0 {
		return domain.Caucion{}, errors.New("taxes must be >= 0")
	}
	if in.Broker == "" {
		in.Broker = "Balanz"
	}

	id, err := newID()
	if err != nil {
		return domain.Caucion{}, err
	}

	tradeDate := in.TradeDate
	if tradeDate.IsZero() {
		tradeDate = time.Now()
	}
	maturity := tradeDate.AddDate(0, 0, in.TermDays)

	gross := ComputeGrossInterest(in.Principal, in.TNA, in.TermDays)

	net := gross.
		Sub(in.Fees).
		Sub(in.Taxes).
		Round(2)

	c := domain.Caucion{
		ID:           id,
		Broker:       in.Broker,
		TradeDate:    tradeDate,
		MaturityDate: maturity,
		TermDays:     in.TermDays,

		Principal:     round2(in.Principal),
		TNA:           round2(in.TNA),
		GrossInterest: round2(gross),

		Fees:        round2(in.Fees),
		Taxes:       round2(in.Taxes),
		NetInterest: round2(net),

		Notes:     in.Notes,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Append(ctx, c); err != nil {
		return domain.Caucion{}, err
	}
	return c, nil
}

func (s *CaucionService) List(ctx context.Context) ([]domain.Caucion, error) {
	return s.repo.List(ctx)
}
