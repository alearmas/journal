package domain

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type FileJSONRepository struct {
	Path string
}

func NewFileJSONRepository(path string) *FileJSONRepository {
	return &FileJSONRepository{Path: path}
}

func (r *FileJSONRepository) List(ctx context.Context) ([]Caucion, error) {
	_ = ctx

	b, err := os.ReadFile(r.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Caucion{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return []Caucion{}, nil
	}

	var items []Caucion
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *FileJSONRepository) Append(ctx context.Context, c Caucion) error {
	_ = ctx

	items, err := r.List(context.Background())
	if err != nil {
		return err
	}

	items = append(items, c)

	if err := os.MkdirAll(filepath.Dir(r.Path), 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.Path, b, 0o644)
}
