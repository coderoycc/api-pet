package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID             uuid.UUID
	Name           string
	DocumentType   string
	DocumentNumber string
	Email          *string
	Phone          *string
	MobilePhone    *string
	Address        *string
	City           *string
	Department     *string
	BusinessName   *string
	TaxCategory    *string
	Notes          *string
	Status         string // active, inactive, suspended
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
