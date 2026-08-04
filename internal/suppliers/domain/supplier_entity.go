package domain

import (
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID           uuid.UUID
	Name         string
	TaxID        string
	ContactEmail string
	ContactPhone string
	Address      string
	Status       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
