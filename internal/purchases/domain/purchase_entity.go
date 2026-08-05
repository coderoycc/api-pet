package domain

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseStatus string

const (
	PurchaseStatusPending   PurchaseStatus = "pending"
	PurchaseStatusConfirmed PurchaseStatus = "confirmed"
	PurchaseStatusReceived  PurchaseStatus = "received"
	PurchaseStatusCancelled PurchaseStatus = "cancelled"
)

type Purchase struct {
	ID          uuid.UUID
	Number      string
	SupplierID  uuid.UUID
	WarehouseID uuid.UUID
	Items       []PurchaseItem
	Status      PurchaseStatus
	TotalAmount float64
	Notes       *string
	CreatedAt   time.Time
	ReceivedAt  *time.Time
	UpdatedAt   time.Time

	// Read-only properties populated via joins
	SupplierName  string
	WarehouseName string
}

type PurchaseItem struct {
	ID         uuid.UUID
	PurchaseID uuid.UUID
	ProductID  uuid.UUID
	Quantity   int
	UnitCost   float64
	Subtotal   float64

	// Read-only properties populated via joins
	ProductName string
	SKU         string
	Unit        string
}

type PurchaseMetrics struct {
	TotalOrders          int64
	PendingOrders        int64
	ReceivedThisMonth    int64
	TotalAmountThisMonth float64
}
