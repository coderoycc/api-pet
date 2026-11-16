package domain

import (
	"context"

	"github.com/google/uuid"
)

type SaleRepository interface {
	Create(ctx context.Context, sale *Sale) error
	GetByID(ctx context.Context, id uuid.UUID) (*Sale, error)
	GetAll(ctx context.Context, filters *SaleFilters) (*PaginatedSaleResult, error)
	GetMetrics(ctx context.Context) (*SaleMetrics, error)
}
