package application

import (
	"github.com/google/uuid"

	"api-go/internal/customers/domain"
)

// CustomerCreateDTO contiene los datos para crear un nuevo cliente.
type CustomerCreateDTO struct {
	Name           string  `json:"name"`
	DocumentType   string  `json:"documentType"`
	DocumentNumber string  `json:"documentNumber"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	MobilePhone    *string `json:"mobilePhone"`
	Address        *string `json:"address"`
	City           *string `json:"city"`
	Department     *string `json:"department"`
	BusinessName   *string `json:"businessName"`
	TaxCategory    *string `json:"taxCategory"`
	Notes          *string `json:"notes"`
	Status         string  `json:"status"` // active, inactive, suspended
}

// CustomerUpdateDTO contiene los datos para actualizar un cliente.
type CustomerUpdateDTO struct {
	Name           string  `json:"name"`
	DocumentType   string  `json:"documentType"`
	DocumentNumber string  `json:"documentNumber"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	MobilePhone    *string `json:"mobilePhone"`
	Address        *string `json:"address"`
	City           *string `json:"city"`
	Department     *string `json:"department"`
	BusinessName   *string `json:"businessName"`
	TaxCategory    *string `json:"taxCategory"`
	Notes          *string `json:"notes"`
	Status         string  `json:"status"`
}

// CustomerPaginationMeta contiene los metadatos de paginación compatibles con el frontend.
type CustomerPaginationMeta struct {
	CurrentPage int   `json:"currentPage"`
	PageSize    int   `json:"pageSize"`
	TotalItems  int64 `json:"totalItems"`
	TotalPages  int   `json:"totalPages"`
}

// CustomerListResponse es la respuesta paginada del listado de clientes.
type CustomerListResponse struct {
	Data       []*domain.Customer     `json:"data"`
	Pagination CustomerPaginationMeta `json:"pagination"`
	Total      int64                  `json:"total"` // compatibilidad dual con el frontend
}

// CustomerSelectorDTO es el DTO optimizado para selectores/autocompletado.
type CustomerSelectorDTO struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	DocumentType   string    `json:"documentType"`
	DocumentNumber string    `json:"documentNumber"`
	BusinessName   *string   `json:"businessName"`
	Email          *string   `json:"email"`
	MobilePhone    *string   `json:"mobilePhone"`
	Status         string    `json:"status"`
}
