package domain

import (
	"context"

	"github.com/google/uuid"
)

type SupplierFilters struct {
	Search string
	Status *bool
}

type SupplierRepository interface {
	Create(ctx context.Context, supplier *Supplier) error
	GetByID(ctx context.Context, id uuid.UUID) (*Supplier, error)
	GetAll(ctx context.Context, filters *SupplierFilters) ([]*Supplier, error)
	Update(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsInUse(ctx context.Context, id uuid.UUID) (bool, error)
}
