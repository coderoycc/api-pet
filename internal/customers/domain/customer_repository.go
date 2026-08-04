package domain

import (
	"context"

	"github.com/google/uuid"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	GetAll(ctx context.Context) ([]*Customer, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
}
