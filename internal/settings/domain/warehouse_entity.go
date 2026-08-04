package domain

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID        uuid.UUID
	Name      string
	Address   string
	Location  string
	Type      string // "store" or "warehouse" etc.
	Status    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
