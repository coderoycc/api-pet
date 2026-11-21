package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Constantes estándar de acciones para mantener uniformidad entre módulos.
const (
	ActionCreate  = "CREATE"
	ActionUpdate  = "UPDATE"
	ActionDelete  = "DELETE"
	ActionReceive = "RECEIVE"
	ActionCancel  = "CANCEL"
	ActionAdjust  = "ADJUST"
	ActionLogin   = "LOGIN"
)

// Actor representa al usuario o sistema que ejecuta la acción auditada.
type Actor struct {
	ID        string
	Email     string
	Role      string
	IPAddress string
	UserAgent string
}

// AuditEntry es el DTO de entrada que cualquier módulo utiliza para registrar un evento.
type AuditEntry struct {
	Module      string
	Action      string
	EntityType  string
	EntityID    string
	Description string
	Before      any
	After       any
	Metadata    map[string]any
	Actor       *Actor // Opcional: si es nil, se extrae automáticamente del context.Context
}

// AuditLog representa un registro persistido en la bitácora del sistema.
type AuditLog struct {
	ID          uuid.UUID       `json:"id"`
	Module      string          `json:"module"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Description string          `json:"description"`
	ActorID     string          `json:"actor_id"`
	ActorEmail  string          `json:"actor_email"`
	ActorRole   string          `json:"actor_role"`
	IPAddress   string          `json:"ip_address"`
	UserAgent   string          `json:"user_agent"`
	OldValues   json.RawMessage `json:"old_values,omitempty"`
	NewValues   json.RawMessage `json:"new_values,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// AuditService es la interfaz pública que consumen los demás módulos del sistema.
type AuditService interface {
	Record(ctx context.Context, entry AuditEntry)
	RecordTx(ctx context.Context, tx *gorm.DB, entry AuditEntry) error
	List(ctx context.Context, filter AuditFilter) (*AuditListResult, error)
}
