package domain_test

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileJSONRepository_ListEmptyFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.json")

	// Create an empty file
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	repo := domain.NewFileJSONRepository(path)
	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestFileJSONRepository_ListNonExistentFile(t *testing.T) {
	repo := domain.NewFileJSONRepository("/nonexistent/path/data.json")
	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error for non-existent file: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestFileJSONRepository_ListInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.json")

	if err := os.WriteFile(path, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := domain.NewFileJSONRepository(path)
	_, err := repo.List(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	var repoErr *domain.ErrRepository
	if !errors.As(err, &repoErr) {
		t.Fatalf("expected *ErrRepository, got %T: %v", err, err)
	}
	if repoErr.Op != "list" {
		t.Fatalf("expected op 'list', got %q", repoErr.Op)
	}
}

func TestFileJSONRepository_AppendCreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sub", "dir", "data.json")

	repo := domain.NewFileJSONRepository(path)
	c := domain.Caucion{ID: "test-1", Broker: "Test"}

	if err := repo.Append(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestFileJSONRepository_ListReadPermissionError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "noperm.json")

	if err := os.WriteFile(path, []byte("[]"), 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	repo := domain.NewFileJSONRepository(path)
	_, err := repo.List(context.Background())
	if err == nil {
		t.Fatal("expected error for unreadable file, got nil")
	}

	var repoErr *domain.ErrRepository
	if !errors.As(err, &repoErr) {
		t.Fatalf("expected *ErrRepository, got %T: %v", err, err)
	}
}
