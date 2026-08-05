package domain

import (
	"context"

	"github.com/google/uuid"
)

type ProductFilters struct {
	Search      string
	Category    string
	Status      string
	MinPrice    *float64
	MaxPrice    *float64
	SupplierID  *uuid.UUID
	Tags        []string
}

type PaginatedResult struct {
	Data       []*Product
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

type ExpiringProductFilters struct {
	DaysAhead *int
	Search    string
	Category  string
	Urgency   string
}

type ExpiringProduct struct {
	ProductID       uuid.UUID
	ProductName     string
	SKU             string
	Category        string
	BatchID         uuid.UUID
	BatchNumber     string
	ExpirationDate  string
	DaysUntilExpiry int
	Quantity        int
	Unit            string
	Urgency         string
	SupplierName    string
}

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetAll(ctx context.Context, page, limit int, filters *ProductFilters, sortBy, sortOrder string) (*PaginatedResult, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	GetProductsByCategory(ctx context.Context, category string) ([]*Product, error)
	GetExpiringProducts(ctx context.Context, filters *ExpiringProductFilters) ([]*ExpiringProduct, int64, error)
	GetLowStockProducts(ctx context.Context, threshold int) ([]*Product, error)
}
