package domain

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type FileJSONRepository struct {
	Path string
	mu   sync.Mutex
}

func NewFileJSONRepository(path string) *FileJSONRepository {
	return &FileJSONRepository{Path: path}
}

func (r *FileJSONRepository) List(ctx context.Context) ([]Caucion, error) {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.listUnsafe()
}

func (r *FileJSONRepository) listUnsafe() ([]Caucion, error) {
	b, err := os.ReadFile(r.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Caucion{}, nil
		}
		return nil, &ErrRepository{Op: "list", Err: err}
	}
	if len(b) == 0 {
		return []Caucion{}, nil
	}

	var items []Caucion
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, &ErrRepository{Op: "list", Err: err}
	}
	return items, nil
}

func (r *FileJSONRepository) Append(ctx context.Context, c Caucion) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := r.listUnsafe()
	if err != nil {
		return err
	}

	items = append(items, c)

	if err := os.MkdirAll(filepath.Dir(r.Path), 0o755); err != nil {
		return &ErrRepository{Op: "append", Err: err}
	}

	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return &ErrRepository{Op: "append", Err: err}
	}
	if err := os.WriteFile(r.Path, b, 0o644); err != nil {
		return &ErrRepository{Op: "append", Err: err}
	}
	return nil
}
