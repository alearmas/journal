package domain

import "fmt"

// ErrValidation represents a validation error for caucion input fields.
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}

// ErrParse represents an error parsing stored data (dates, decimals, etc.).
type ErrParse struct {
	Field string
	Value string
	Err   error
}

func (e *ErrParse) Error() string {
	return fmt.Sprintf("parse %s=%q: %v", e.Field, e.Value, e.Err)
}

func (e *ErrParse) Unwrap() error { return e.Err }

// ErrRepository represents a generic repository operation error.
type ErrRepository struct {
	Op  string // e.g. "append", "list", "init"
	Err error
}

func (e *ErrRepository) Error() string {
	return fmt.Sprintf("repository %s: %v", e.Op, e.Err)
}

func (e *ErrRepository) Unwrap() error { return e.Err }
