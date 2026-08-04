package domain

import (
	"context"

	"github.com/google/uuid"
)

type UnitOfMeasureRepository interface {
	Create(ctx context.Context, uom *UnitOfMeasure) error
	GetByID(ctx context.Context, id uuid.UUID) (*UnitOfMeasure, error)
	GetAll(ctx context.Context) ([]*UnitOfMeasure, error)
	Update(ctx context.Context, uom *UnitOfMeasure) error
	Delete(ctx context.Context, id uuid.UUID) error
}
