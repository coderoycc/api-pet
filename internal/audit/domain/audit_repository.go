package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// AuditFilter define los criterios de búsqueda y paginación para consultar la bitácora.
type AuditFilter struct {
	Module     string
	Action     string
	EntityType string
	EntityID   string
	ActorID    string
	From       *time.Time
	To         *time.Time
	Page       int
	Limit      int
}

// AuditListResult representa el resultado paginado de la consulta de bitácora.
type AuditListResult struct {
	Items      []AuditLog `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}

// AuditRepository define el contrato de persistencia para los registros de auditoría.
type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	CreateTx(ctx context.Context, tx *gorm.DB, log *AuditLog) error
	FindAll(ctx context.Context, filter AuditFilter) (*AuditListResult, error)
}
