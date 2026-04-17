// Package filejson provides a file-based JSON driven adapter for the repository ports.
package filejson

import (
	"alearmas/tradingJournal/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// FileJSONRepository persists cauciones as a JSON array in a single file.
type FileJSONRepository struct {
	Path string
	mu   sync.Mutex
}

// NewFileJSONRepository creates a new FileJSONRepository for the given file path.
func NewFileJSONRepository(path string) *FileJSONRepository {
	return &FileJSONRepository{Path: path}
}

func (r *FileJSONRepository) List(ctx context.Context) ([]domain.Caucion, error) {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.listUnsafe()
}

func (r *FileJSONRepository) listUnsafe() ([]domain.Caucion, error) {
	b, err := os.ReadFile(r.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []domain.Caucion{}, nil
		}
		return nil, &domain.ErrRepository{Op: "list", Err: err}
	}
	if len(b) == 0 {
		return []domain.Caucion{}, nil
	}

	var items []domain.Caucion
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, &domain.ErrRepository{Op: "list", Err: err}
	}
	return items, nil
}

func (r *FileJSONRepository) Append(ctx context.Context, c domain.Caucion) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := r.listUnsafe()
	if err != nil {
		return err
	}

	items = append(items, c)

	if err := os.MkdirAll(filepath.Dir(r.Path), 0o755); err != nil {
		return &domain.ErrRepository{Op: "append", Err: err}
	}

	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return &domain.ErrRepository{Op: "append", Err: err}
	}
	if err := os.WriteFile(r.Path, b, 0o644); err != nil {
		return &domain.ErrRepository{Op: "append", Err: err}
	}
	return nil
}
