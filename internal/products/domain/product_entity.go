package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID
	SKU           string
	Name          string
	Description   string
	Category      string
	Subcategory   string
	Price         float64
	Cost          float64
	Quantity      float64
	MinQuantity   float64
	MaxQuantity   *float64
	Unit          string
	Status        string
	SupplierID    *uuid.UUID
	Tags          []string
	Images        []ProductImage
	Batches       []ProductBatch
	Presentations []ProductPresentation
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ProductPresentation struct {
	ID               uuid.UUID
	ProductID        uuid.UUID
	Name             string
	Barcode          string
	ConversionFactor float64
	SalePrice        float64
	IsDefault        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ProductImage struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	URL       string
	Alt       string
	IsPrimary bool
}

type ProductBatch struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	BatchNumber    string
	ExpirationDate time.Time
	Quantity       float64
	SupplierID     *uuid.UUID
}
