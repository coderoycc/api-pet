package application

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"api-go/internal/audit/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditService struct {
	repo domain.AuditRepository
}

// NewAuditService construye el servicio genérico de auditoría.
func NewAuditService(repo domain.AuditRepository) domain.AuditService {
	return &auditService{repo: repo}
}

// Record registra un evento de auditoría de forma no bloqueante para el flujo de negocio:
// si ocurre un error al persistir, se registra en log sin abortar la operación principal.
func (s *auditService) Record(ctx context.Context, entry domain.AuditEntry) {
	logEntry := s.buildLogEntry(ctx, entry)
	if err := s.repo.Create(ctx, logEntry); err != nil {
		log.Printf("[audit] warning: failed to record audit log (%s/%s): %v", entry.Module, entry.Action, err)
	}
}

// RecordTx registra un evento de auditoría dentro de una transacción GORM existente.
func (s *auditService) RecordTx(ctx context.Context, tx *gorm.DB, entry domain.AuditEntry) error {
	logEntry := s.buildLogEntry(ctx, entry)
	return s.repo.CreateTx(ctx, tx, logEntry)
}

// List consulta la bitácora aplicando los filtros y paginación especificados.
func (s *auditService) List(ctx context.Context, filter domain.AuditFilter) (*domain.AuditListResult, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *auditService) buildLogEntry(ctx context.Context, entry domain.AuditEntry) *domain.AuditLog {
	actor := resolveActor(ctx, entry.Actor)

	return &domain.AuditLog{
		ID:          uuid.New(),
		Module:      entry.Module,
		Action:      entry.Action,
		EntityType:  entry.EntityType,
		EntityID:    entry.EntityID,
		Description: entry.Description,
		ActorID:     actor.ID,
		ActorEmail:  actor.Email,
		ActorRole:   actor.Role,
		IPAddress:   truncateString(actor.IPAddress, 45),
		UserAgent:   truncateString(actor.UserAgent, 255),
		OldValues:   marshalOptionalJSON(entry.Before),
		NewValues:   marshalOptionalJSON(entry.After),
		Metadata:    marshalOptionalJSON(entry.Metadata),
		CreatedAt:   time.Now().UTC(),
	}
}

func resolveActor(ctx context.Context, explicit *domain.Actor) domain.Actor {
	if explicit != nil {
		actor := *explicit
		if actor.ID == "" {
			actor.ID = "system"
		}
		if actor.Email == "" {
			actor.Email = actor.ID
		}
		return actor
	}

	if fromCtx, ok := ActorFromContext(ctx); ok && fromCtx.ID != "" {
		if fromCtx.Email == "" {
			fromCtx.Email = fromCtx.ID
		}
		return fromCtx
	}

	return domain.Actor{
		ID:    "system",
		Email: "system",
		Role:  "system",
	}
}

func marshalOptionalJSON(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("[audit] warning: failed to marshal json payload: %v", err)
		return nil
	}
	if string(data) == "null" {
		return nil
	}
	return data
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
