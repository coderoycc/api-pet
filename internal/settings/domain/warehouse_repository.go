package domain

import (
	"context"

	"github.com/google/uuid"
)

type WarehouseRepository interface {
	Create(ctx context.Context, w *Warehouse) error
	GetByID(ctx context.Context, id uuid.UUID) (*Warehouse, error)
	GetAll(ctx context.Context) ([]*Warehouse, error)
	Update(ctx context.Context, w *Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error
}
