package domain

import "context"

type CaucionRepository interface {
	List(ctx context.Context) ([]Caucion, error)
	Append(ctx context.Context, c Caucion) error
}
