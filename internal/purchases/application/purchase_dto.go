package application

import (
	"api-go/internal/purchases/domain"
	"time"
)

type PurchaseItemDto struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitCost  float64 `json:"unitCost"`
}

type PurchaseCreateDto struct {
	SupplierID  string            `json:"supplierId"`
	WarehouseID string            `json:"warehouseId"`
	Items       []PurchaseItemDto `json:"items"`
	Notes       *string           `json:"notes,omitempty"`
}

type PurchaseItemResponse struct {
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitCost    float64 `json:"unitCost"`
	Unit        string  `json:"unit"`
	Subtotal    float64 `json:"subtotal"`
}

type PurchaseResponse struct {
	ID            string                 `json:"id"`
	Number        string                 `json:"number"`
	SupplierID    string                 `json:"supplierId"`
	SupplierName  string                 `json:"supplierName"`
	WarehouseID   string                 `json:"warehouseId"`
	WarehouseName string                 `json:"warehouseName"`
	Items         []PurchaseItemResponse `json:"items"`
	Status        domain.PurchaseStatus  `json:"status"`
	TotalAmount   float64                `json:"totalAmount"`
	Notes         *string                `json:"notes,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	ReceivedAt    *time.Time             `json:"receivedAt,omitempty"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

type PurchaseListResponse struct {
	Data       []*PurchaseResponse `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"totalPages"`
}

type PurchaseMetricsResponse struct {
	TotalOrders          int64   `json:"totalOrders"`
	PendingOrders        int64   `json:"pendingOrders"`
	ReceivedThisMonth    int64   `json:"receivedThisMonth"`
	TotalAmountThisMonth float64 `json:"totalAmountThisMonth"`
}

type BaseResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
