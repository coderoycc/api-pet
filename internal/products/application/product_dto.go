package application

import (
	"api-go/internal/products/domain"
	"time"
)

type ProductImageDto struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	IsPrimary bool   `json:"isPrimary"`
}

type ProductCreateImageDto struct {
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	IsPrimary bool   `json:"isPrimary"`
}

type ProductBatchDto struct {
	ID             string    `json:"id"`
	BatchNumber    string    `json:"batchNumber"`
	ExpirationDate time.Time `json:"expirationDate"`
	Quantity       int       `json:"quantity"`
	SupplierID     *string   `json:"supplierId,omitempty"`
}

type ProductCreateDto struct {
	SKU         string                  `json:"sku"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Category    string                  `json:"category"`
	Subcategory string                  `json:"subcategory,omitempty"`
	Price       float64                 `json:"price"`
	Cost        float64                 `json:"cost"`
	Quantity    int                     `json:"quantity"`
	MinQuantity int                     `json:"minQuantity"`
	MaxQuantity *int                    `json:"maxQuantity,omitempty"`
	Unit        string                  `json:"unit"`
	Status      string                  `json:"status"`
	SupplierID  *string                 `json:"supplierId,omitempty"`
	Images      []ProductCreateImageDto `json:"images,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
}

type ProductUpdateDto struct {
	ID          string                  `json:"id"`
	SKU         *string                 `json:"sku,omitempty"`
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Category    *string                 `json:"category,omitempty"`
	Subcategory *string                 `json:"subcategory,omitempty"`
	Price       *float64                `json:"price,omitempty"`
	Cost        *float64                `json:"cost,omitempty"`
	Quantity    *int                    `json:"quantity,omitempty"`
	MinQuantity *int                    `json:"minQuantity,omitempty"`
	MaxQuantity *int                    `json:"maxQuantity,omitempty"`
	Unit        *string                 `json:"unit,omitempty"`
	Status      *string                 `json:"status,omitempty"`
	SupplierID  *string                 `json:"supplierId,omitempty"`
	Images      []ProductCreateImageDto `json:"images,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
}

type ProductResponse struct {
	ID          string            `json:"id"`
	SKU         string            `json:"sku"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Subcategory string            `json:"subcategory,omitempty"`
	Price       float64           `json:"price"`
	Cost        float64           `json:"cost"`
	Quantity    int               `json:"quantity"`
	MinQuantity int               `json:"minQuantity"`
	MaxQuantity *int              `json:"maxQuantity,omitempty"`
	Unit        string            `json:"unit"`
	Status      string            `json:"status"`
	SupplierID  *string           `json:"supplierId,omitempty"`
	Images      []ProductImageDto `json:"images,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type ProductListResponse struct {
	Data       []*ProductResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
}

type ExpiringProductsResponse struct {
	Data  []*domain.ExpiringProduct `json:"data"`
	Total int64                     `json:"total"`
}

type BaseResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}
