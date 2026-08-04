package application

import "time"

type CreateSupplierDTO struct {
	Name         string `json:"name"`
	TaxID        string `json:"taxId"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	Address      string `json:"address"`
	Status       bool   `json:"status"`
}

type UpdateSupplierDTO struct {
	Name         string `json:"name"`
	TaxID        string `json:"taxId"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	Address      string `json:"address"`
	Status       bool   `json:"status"`
}

type SupplierResponseDTO struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	TaxID        string    `json:"taxId"`
	ContactEmail string    `json:"contactEmail"`
	ContactPhone string    `json:"contactPhone"`
	Address      string    `json:"address"`
	Status       bool      `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
