package domain

import (
	"context"

	"github.com/google/uuid"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	GetAll(ctx context.Context, page, limit int, filters *CustomerFilters, sortBy string, descending bool) (*PaginatedCustomerResult, error)
	Search(ctx context.Context, query string, limit int, status string) ([]*Customer, error)
	GetDistinctCities(ctx context.Context) ([]string, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckDocumentConflict(ctx context.Context, docType, docNumber string, excludeID *uuid.UUID) (bool, error)
	HasAssociatedRecords(ctx context.Context, id uuid.UUID) (bool, error)
}

