package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"api-go/internal/audit/domain"

	"gorm.io/gorm"
)

type mockAuditRepo struct {
	createdLogs []*domain.AuditLog
	lastTx      *gorm.DB
	createErr   error
	findResult  *domain.AuditListResult
}

func (m *mockAuditRepo) Create(ctx context.Context, logEntry *domain.AuditLog) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createdLogs = append(m.createdLogs, logEntry)
	return nil
}

func (m *mockAuditRepo) CreateTx(ctx context.Context, tx *gorm.DB, logEntry *domain.AuditLog) error {
	m.lastTx = tx
	if m.createErr != nil {
		return m.createErr
	}
	m.createdLogs = append(m.createdLogs, logEntry)
	return nil
}

func (m *mockAuditRepo) FindAll(ctx context.Context, filter domain.AuditFilter) (*domain.AuditListResult, error) {
	return m.findResult, nil
}

func TestAuditService_Record_ActorFromContextAndPayloads(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo)

	ctx := WithActor(context.Background(), domain.Actor{
		ID:        "user-123",
		Email:     "admin@pet.com",
		Role:      "Administrador",
		IPAddress: "127.0.0.1",
		UserAgent: "GoTest/1.0",
	})

	svc.Record(ctx, domain.AuditEntry{
		Module:      "purchases",
		Action:      domain.ActionCreate,
		EntityType:  "purchase",
		EntityID:    "pur-1",
		Description: "Compra creada",
		Before:      nil,
		After:       map[string]any{"status": "pending"},
		Metadata:    map[string]any{"total": 150.5},
	})

	if len(repo.createdLogs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(repo.createdLogs))
	}

	got := repo.createdLogs[0]
	if got.ActorID != "user-123" || got.ActorEmail != "admin@pet.com" || got.ActorRole != "Administrador" {
		t.Fatalf("unexpected actor fields: %+v", got)
	}
	if got.IPAddress != "127.0.0.1" || got.UserAgent != "GoTest/1.0" {
		t.Fatalf("unexpected request metadata: ip=%s ua=%s", got.IPAddress, got.UserAgent)
	}
	if got.OldValues != nil {
		t.Fatalf("expected nil OldValues, got %s", string(got.OldValues))
	}

	var after map[string]string
	if err := json.Unmarshal(got.NewValues, &after); err != nil || after["status"] != "pending" {
		t.Fatalf("unexpected NewValues: %s, err=%v", string(got.NewValues), err)
	}
}

func TestAuditService_Record_SystemFallbackAndBestEffort(t *testing.T) {
	repo := &mockAuditRepo{createErr: errors.New("db unavailable")}
	svc := NewAuditService(repo)

	// No debe entrar en pánico ni bloquear aunque falle la BD
	svc.Record(context.Background(), domain.AuditEntry{
		Module:      "inventory",
		Action:      domain.ActionAdjust,
		EntityType:  "product",
		EntityID:    "prod-1",
		Description: "Ajuste automático",
	})

	// Ahora sin error, verificamos el fallback a system
	repo.createErr = nil
	svc.Record(context.Background(), domain.AuditEntry{
		Module:      "inventory",
		Action:      domain.ActionAdjust,
		EntityType:  "product",
		EntityID:    "prod-1",
		Description: "Ajuste automático",
	})

	if len(repo.createdLogs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(repo.createdLogs))
	}
	if repo.createdLogs[0].ActorID != "system" || repo.createdLogs[0].ActorRole != "system" {
		t.Fatalf("expected system actor fallback, got %+v", repo.createdLogs[0])
	}
}

func TestAuditService_RecordTx_PropagatesError(t *testing.T) {
	expectedErr := errors.New("tx failed")
	repo := &mockAuditRepo{createErr: expectedErr}
	svc := NewAuditService(repo)

	fakeTx := &gorm.DB{}
	err := svc.RecordTx(context.Background(), fakeTx, domain.AuditEntry{
		Module:      "sales",
		Action:      domain.ActionCreate,
		EntityType:  "sale",
		EntityID:    "sale-1",
		Description: "Venta registrada",
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if repo.lastTx != fakeTx {
		t.Fatalf("expected fakeTx to be passed to repository")
	}
}
