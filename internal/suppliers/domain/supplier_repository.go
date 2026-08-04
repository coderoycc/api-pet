package domain

import (
	"context"

	"github.com/google/uuid"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *Supplier) error
	GetByID(ctx context.Context, id uuid.UUID) (*Supplier, error)
	GetAll(ctx context.Context) ([]*Supplier, error)
	Update(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
}
