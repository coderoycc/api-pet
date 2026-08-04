package domain

import (
	"time"

	"github.com/google/uuid"
)

type UnitOfMeasure struct {
	ID           uuid.UUID
	Detail       string
	Abbreviation string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
