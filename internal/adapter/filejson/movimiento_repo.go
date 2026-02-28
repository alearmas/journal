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

// MovimientoFileJSONRepository persists movimientos as a JSON array in a single file.
type MovimientoFileJSONRepository struct {
	Path string
	mu   sync.Mutex
}

// NewMovimientoFileJSONRepository creates a new MovimientoFileJSONRepository.
func NewMovimientoFileJSONRepository(path string) *MovimientoFileJSONRepository {
	return &MovimientoFileJSONRepository{Path: path}
}

func (r *MovimientoFileJSONRepository) List(ctx context.Context) ([]domain.Movimiento, error) {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.listUnsafe()
}

func (r *MovimientoFileJSONRepository) listUnsafe() ([]domain.Movimiento, error) {
	b, err := os.ReadFile(r.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []domain.Movimiento{}, nil
		}
		return nil, &domain.ErrRepository{Op: "list", Err: err}
	}
	if len(b) == 0 {
		return []domain.Movimiento{}, nil
	}

	var items []domain.Movimiento
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, &domain.ErrRepository{Op: "list", Err: err}
	}
	return items, nil
}

func (r *MovimientoFileJSONRepository) Append(ctx context.Context, m domain.Movimiento) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := r.listUnsafe()
	if err != nil {
		return err
	}

	items = append(items, m)

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
