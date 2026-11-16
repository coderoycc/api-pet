package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSaleNotFound      = errors.New("venta no encontrada")
	ErrInsufficientStock = errors.New("stock insuficiente para completar la venta")
	ErrEmptyCart         = errors.New("el carrito de venta no puede estar vacío")
	ErrInvalidPayment    = errors.New("el monto pagado es inferior al total de la venta")
)

type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodQR           PaymentMethod = "qr"
	PaymentMethodTransfer     PaymentMethod = "transfer"
)

type Sale struct {
	ID            uuid.UUID
	InvoiceNumber string
	CustomerID    *uuid.UUID
	CustomerName  string
	UserID        uuid.UUID
	UserName      string
	TotalAmount   float64
	PaymentMethod PaymentMethod
	AmountPaid    float64
	ChangeAmount  float64
	Notes         string
	Items         []SaleItem
	CreatedAt     time.Time
}

type SaleItem struct {
	ID               uuid.UUID
	SaleID           uuid.UUID
	ProductID        uuid.UUID
	ProductName      string
	ProductSKU       string
	ProductUnit      string
	PresentationID   *uuid.UUID
	PresentationName string
	Quantity         float64 // Cantidad vendida en la presentación seleccionada
	ConversionFactor float64 // Factor hacia la unidad base física
	BaseQuantity     float64 // Quantity * ConversionFactor (descontado de stock)
	UnitPrice        float64 // Precio cobrado por unidad de la presentación
	Subtotal         float64 // Quantity * UnitPrice
}

type SaleMetrics struct {
	TodaySalesCount   int64   `json:"todaySalesCount"`
	TodayTotalRevenue float64 `json:"todayTotalRevenue"`
	MonthSalesCount   int64   `json:"monthSalesCount"`
	MonthTotalRevenue float64 `json:"monthTotalRevenue"`
	AverageTicket     float64 `json:"averageTicket"`
}

type SaleFilters struct {
	Search        string
	PaymentMethod string
	CustomerID    *uuid.UUID
	UserID        *uuid.UUID
	StartDate     *time.Time
	EndDate       *time.Time
	Page          int
	Limit         int
	SortBy        string
	SortOrder     string
}

type PaginatedSaleResult struct {
	Data       []*Sale
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}
