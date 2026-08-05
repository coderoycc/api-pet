package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	SKU         string
	Name        string
	Description string
	Category    string
	Subcategory string
	Price       float64
	Cost        float64
	Quantity    int
	MinQuantity int
	MaxQuantity *int
	Unit        string
	Status      string
	SupplierID  *uuid.UUID
	Tags        []string
	Images      []ProductImage
	Batches     []ProductBatch
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	Quantity       int
	SupplierID     *uuid.UUID
}
