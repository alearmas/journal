package export_test

import (
	"alearmas/tradingJournal/internal/domain"
	"alearmas/tradingJournal/internal/export"
	"path/filepath"
	"testing"
)

func TestWriteCaucionesCSV_EmptyList(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "empty.csv")

	if err := export.WriteCaucionesCSV(out, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteCaucionesCSV_InvalidPath(t *testing.T) {
	err := export.WriteCaucionesCSV("/nonexistent/deep/path/out.csv", []domain.Caucion{})
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}
