package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	DocumentType   string    `json:"documentType"`
	DocumentNumber string    `json:"documentNumber"`
	Email          *string   `json:"email"`
	Phone          *string   `json:"phone"`
	MobilePhone    *string   `json:"mobilePhone"`
	Address        *string   `json:"address"`
	City           *string   `json:"city"`
	Department     *string   `json:"department"`
	BusinessName   *string   `json:"businessName"`
	TaxCategory    *string   `json:"taxCategory"`
	Notes          *string   `json:"notes"`
	Status         string    `json:"status"` // active, inactive, suspended
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CustomerFilters contiene los criterios de filtrado para la consulta paginada.
type CustomerFilters struct {
	Search        string   // Búsqueda global en name, documentNumber, businessName, email, phone, mobilePhone
	DocumentTypes []string // nit | ci | passport | other
	Statuses      []string // active | inactive | suspended
	City          string
}

// PaginatedCustomerResult contiene los resultados paginados de clientes.
type PaginatedCustomerResult struct {
	Data       []*Customer
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

