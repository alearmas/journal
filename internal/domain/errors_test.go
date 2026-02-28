package domain_test

import (
	"alearmas/tradingJournal/internal/domain"
	"errors"
	"fmt"
	"testing"
)

func TestErrValidation_Error(t *testing.T) {
	e := &domain.ErrValidation{Field: "principal", Message: "must be > 0"}
	want := "validation: principal: must be > 0"
	if e.Error() != want {
		t.Fatalf("got %q, want %q", e.Error(), want)
	}
}

func TestErrParse_Error(t *testing.T) {
	inner := fmt.Errorf("bad format")
	e := &domain.ErrParse{Field: "trade_date", Value: "not-a-date", Err: inner}
	got := e.Error()
	if got != `parse trade_date="not-a-date": bad format` {
		t.Fatalf("unexpected message: %s", got)
	}
}

func TestErrParse_Unwrap(t *testing.T) {
	inner := fmt.Errorf("original")
	e := &domain.ErrParse{Field: "f", Value: "v", Err: inner}
	if !errors.Is(e, inner) {
		t.Fatal("Unwrap should return inner error")
	}
}

func TestErrRepository_Error(t *testing.T) {
	inner := fmt.Errorf("disk full")
	e := &domain.ErrRepository{Op: "append", Err: inner}
	want := "repository append: disk full"
	if e.Error() != want {
		t.Fatalf("got %q, want %q", e.Error(), want)
	}
}

func TestErrRepository_Unwrap(t *testing.T) {
	inner := fmt.Errorf("original")
	e := &domain.ErrRepository{Op: "list", Err: inner}
	if !errors.Is(e, inner) {
		t.Fatal("Unwrap should return inner error")
	}
}
