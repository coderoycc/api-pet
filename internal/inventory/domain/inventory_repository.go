package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InventoryFilters struct {
	Type       *InventoryLogType `json:"type,omitempty"`
	ProductID  *uuid.UUID        `json:"productId,omitempty"`
	BatchID    *uuid.UUID        `json:"batchId,omitempty"`
	UserID     *string           `json:"userId,omitempty"`
	ReasonType *string           `json:"reasonType,omitempty"`
	StartDate  *time.Time        `json:"startDate,omitempty"`
	EndDate    *time.Time        `json:"endDate,omitempty"`
	Search     string            `json:"search,omitempty"`
	Page       int               `json:"page,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	SortBy     string            `json:"sortBy,omitempty"`
	SortOrder  string            `json:"sortOrder,omitempty"`
}

type PaginatedLogsResult struct {
	Data       []*InventoryLog `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"totalPages"`
}

type InventoryRepository interface {
	GetBatchesByProductID(ctx context.Context, productID uuid.UUID) ([]*InventoryBatch, error)
	GetAllBatches(ctx context.Context) ([]*InventoryBatch, error)
	GetBatchByID(ctx context.Context, id uuid.UUID) (*InventoryBatch, error)
	CreateBatch(ctx context.Context, batch *InventoryBatch) error
	UpdateBatchQuantity(ctx context.Context, id uuid.UUID, quantity int) error
	GetAvailableBatchesForProduct(ctx context.Context, productID uuid.UUID) ([]*InventoryBatch, error)

	CreateLog(ctx context.Context, log *InventoryLog) error
	GetLogByID(ctx context.Context, id uuid.UUID) (*InventoryLog, error)
	UpdateLogUndoable(ctx context.Context, id uuid.UUID, undoable bool) error
	GetLogs(ctx context.Context, filters *InventoryFilters) (*PaginatedLogsResult, error)

	CreateStockAdjustment(ctx context.Context, adj *StockAdjustment) error
	GetCriticalStock(ctx context.Context, minStock int) ([]*CriticalStockItem, error)
	GetMetrics(ctx context.Context) (*InventoryMetrics, error)
}
