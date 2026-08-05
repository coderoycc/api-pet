package domain

import (
	"time"

	"github.com/google/uuid"
)

type InventoryLogType string

const (
	LogTypeInbound        InventoryLogType = "inbound"
	LogTypeOutbound       InventoryLogType = "outbound"
	LogTypeAdjustment     InventoryLogType = "adjustment"
	LogTypeManual         InventoryLogType = "manual"
	LogTypeSupplierReturn InventoryLogType = "supplier_return"
)

type InventoryBatch struct {
	ID             uuid.UUID  `json:"id"`
	ProductID      uuid.UUID  `json:"productId"`
	BatchNumber    string     `json:"batchNumber"`
	ExpirationDate *time.Time `json:"expirationDate,omitempty"`
	Quantity       int        `json:"quantity"`
	Cost           *float64   `json:"cost,omitempty"`
	SupplierID     *uuid.UUID `json:"supplierId,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type InventoryLog struct {
	ID            uuid.UUID        `json:"id"`
	ProductID     uuid.UUID        `json:"productId"`
	ProductName   string           `json:"productName"`
	VariantID     *uuid.UUID       `json:"variantId,omitempty"`
	VariantName   *string          `json:"variantName,omitempty"`
	BatchID       *uuid.UUID       `json:"batchId,omitempty"`
	BatchNumber   *string          `json:"batchNumber,omitempty"`
	SKU           string           `json:"sku"`
	Type          InventoryLogType `json:"type"`
	Quantity      int              `json:"quantity"`
	PreviousStock int              `json:"previousStock"`
	NewStock      int              `json:"newStock"`
	Reason        string           `json:"reason"`
	ReasonType    *string          `json:"reasonType,omitempty"`
	UserID        string           `json:"userId"`
	UserName      string           `json:"userName"`
	Notes         *string          `json:"notes,omitempty"`
	CreatedAt     time.Time        `json:"createdAt"`
	Undoable      bool             `json:"undoable"`
}

type StockAdjustment struct {
	ID          uuid.UUID  `json:"id"`
	ProductID   uuid.UUID  `json:"productId"`
	ProductName string     `json:"productName"`
	BatchID     *uuid.UUID `json:"batchId,omitempty"`
	BatchNumber *string    `json:"batchNumber,omitempty"`
	SKU         string     `json:"sku"`
	OldStock    int        `json:"oldStock"`
	NewStock    int        `json:"newStock"`
	Reason      string     `json:"reason"`
	UserID      string     `json:"userId"`
	UserName    string     `json:"userName"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type CriticalStockItem struct {
	ProductID        uuid.UUID  `json:"productId"`
	ProductName      string     `json:"productName"`
	VariantID        *uuid.UUID `json:"variantId,omitempty"`
	VariantName      *string    `json:"variantName,omitempty"`
	SKU              string     `json:"sku"`
	CurrentStock     int        `json:"currentStock"`
	MinStock         int        `json:"minStock"`
	CategoryName     *string    `json:"categoryName,omitempty"`
	BrandName        *string    `json:"brandName,omitempty"`
	AvgDailyUsage    *float64   `json:"avgDailyUsage,omitempty"`
	EstimatedDaysOut *int       `json:"estimatedDaysOut,omitempty"`
}

type InventoryMetrics struct {
	TotalValue           float64 `json:"totalValue"`
	TotalProducts        int64   `json:"totalProducts"`
	CriticalItems        int64   `json:"criticalItems"`
	OutOfStock           int64   `json:"outOfStock"`
	InboundThisMonth     int64   `json:"inboundThisMonth"`
	OutboundThisMonth    int64   `json:"outboundThisMonth"`
	AdjustmentsThisMonth int64   `json:"adjustmentsThisMonth"`
}

type StockOperation struct {
	ProductID  uuid.UUID  `json:"productId"`
	VariantID  *uuid.UUID `json:"variantId,omitempty"`
	BatchID    *uuid.UUID `json:"batchId,omitempty"`
	Quantity   int        `json:"quantity"`
	Reason     string     `json:"reason"`
	ReasonType *string    `json:"reasonType,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
}

type InboundOperation struct {
	StockOperation
	BatchNumber    *string    `json:"batchNumber,omitempty"`
	ExpirationDate *time.Time `json:"expirationDate,omitempty"`
	SupplierName   *string    `json:"supplierName,omitempty"`
	InvoiceNumber  *string    `json:"invoiceNumber,omitempty"`
	Cost           *float64   `json:"cost,omitempty"`
	EstimatedCost  *float64   `json:"estimatedCost,omitempty"`
}

type OutboundOperation struct {
	StockOperation
	CustomerID      *uuid.UUID `json:"customerId,omitempty"`
	SaleID          *uuid.UUID `json:"saleId,omitempty"`
	ReferenceNumber *string    `json:"referenceNumber,omitempty"`
}

type SupplierReturnOperation struct {
	ProductID    uuid.UUID  `json:"productId"`
	VariantID    *uuid.UUID `json:"variantId,omitempty"`
	BatchID      *uuid.UUID `json:"batchId,omitempty"`
	Quantity     int        `json:"quantity"`
	SupplierID   uuid.UUID  `json:"supplierId"`
	ReturnType   string     `json:"returnType"`   // monetary | replacement
	ReasonType   string     `json:"reasonType"`   // defective | expired | excess | other
	Reason       string     `json:"reason"`
	RefundAmount *float64   `json:"refundAmount,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

type QuickAdjustment struct {
	ProductID uuid.UUID  `json:"productId"`
	VariantID *uuid.UUID `json:"variantId,omitempty"`
	BatchID   *uuid.UUID `json:"batchId,omitempty"`
	NewStock  int        `json:"newStock"`
	Reason    string     `json:"reason"`
}

type AdjustStockRequest struct {
	ProductID uuid.UUID  `json:"productId"`
	BatchID   *uuid.UUID `json:"batchId,omitempty"`
	NewStock  int        `json:"newStock"`
	Reason    string     `json:"reason"`
}
