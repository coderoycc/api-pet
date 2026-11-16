package application

import (
	"time"
)

type SaleItemCreateDto struct {
	ProductID        string  `json:"productId"`
	PresentationID   *string `json:"presentationId,omitempty"`
	PresentationName string  `json:"presentationName,omitempty"`
	Quantity         float64 `json:"quantity"`
	ConversionFactor float64 `json:"conversionFactor"`
	UnitPrice        float64 `json:"unitPrice"`
}

type SaleCreateDto struct {
	CustomerID    *string             `json:"customerId,omitempty"`
	CustomerName  string              `json:"customerName,omitempty"`
	PaymentMethod string              `json:"paymentMethod"`
	AmountPaid    float64             `json:"amountPaid"`
	Notes         string              `json:"notes,omitempty"`
	Items         []SaleItemCreateDto `json:"items"`
}

type SaleItemResponse struct {
	ID               string  `json:"id"`
	ProductID        string  `json:"productId"`
	ProductName      string  `json:"productName"`
	ProductSKU       string  `json:"productSku"`
	ProductUnit      string  `json:"productUnit"`
	PresentationID   *string `json:"presentationId,omitempty"`
	PresentationName string  `json:"presentationName,omitempty"`
	Quantity         float64 `json:"quantity"`
	ConversionFactor float64 `json:"conversionFactor"`
	BaseQuantity     float64 `json:"baseQuantity"`
	UnitPrice        float64 `json:"unitPrice"`
	Subtotal         float64 `json:"subtotal"`
}

type SaleResponse struct {
	ID            string             `json:"id"`
	InvoiceNumber string             `json:"invoiceNumber"`
	CustomerID    *string            `json:"customerId,omitempty"`
	CustomerName  string             `json:"customerName"`
	UserID        string             `json:"userId"`
	UserName      string             `json:"userName"`
	TotalAmount   float64            `json:"totalAmount"`
	PaymentMethod string             `json:"paymentMethod"`
	AmountPaid    float64            `json:"amountPaid"`
	ChangeAmount  float64            `json:"changeAmount"`
	Notes         string             `json:"notes"`
	Items         []SaleItemResponse `json:"items"`
	CreatedAt     time.Time          `json:"createdAt"`
}

type SaleMetricsResponse struct {
	TodaySalesCount   int64   `json:"todaySalesCount"`
	TodayTotalRevenue float64 `json:"todayTotalRevenue"`
	MonthSalesCount   int64   `json:"monthSalesCount"`
	MonthTotalRevenue float64 `json:"monthTotalRevenue"`
	AverageTicket     float64 `json:"averageTicket"`
}

type SaleListResponse struct {
	Data       []*SaleResponse `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"totalPages"`
}

type BaseResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
