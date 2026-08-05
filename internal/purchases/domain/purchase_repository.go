package domain

import (
	"context"

	"github.com/google/uuid"
)

type PurchaseFilters struct {
	Search      string
	Status      string
	SupplierID  *uuid.UUID
	WarehouseID *uuid.UUID
	StartDate   string
	EndDate     string
}

type PaginatedPurchaseResult struct {
	Data       []*Purchase
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

type PurchaseRepository interface {
	Create(ctx context.Context, purchase *Purchase) error
	GetByID(ctx context.Context, id uuid.UUID) (*Purchase, error)
	GetAll(ctx context.Context, page, limit int, filters *PurchaseFilters, sortBy, sortOrder string) (*PaginatedPurchaseResult, error)
	UpdateStatus(ctx context.Context, purchase *Purchase) error
	GetMetrics(ctx context.Context) (*PurchaseMetrics, error)
}
