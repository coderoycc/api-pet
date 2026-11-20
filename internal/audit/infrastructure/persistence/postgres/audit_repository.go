package postgres

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"api-go/internal/audit/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditLogModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Module      string    `gorm:"type:varchar(50);not null;index"`
	Action      string    `gorm:"type:varchar(50);not null;index"`
	EntityType  string    `gorm:"type:varchar(50);not null;index:idx_audit_entity"`
	EntityID    string    `gorm:"type:varchar(100);not null;index:idx_audit_entity"`
	Description string    `gorm:"type:text;not null"`
	ActorID     string    `gorm:"type:varchar(100);not null;index"`
	ActorEmail  string    `gorm:"type:varchar(150);not null"`
	ActorRole   string    `gorm:"type:varchar(50);not null"`
	IPAddress   string    `gorm:"type:varchar(45)"`
	UserAgent   string    `gorm:"type:varchar(255)"`
	OldValues   []byte    `gorm:"type:jsonb"`
	NewValues   []byte    `gorm:"type:jsonb"`
	Metadata    []byte    `gorm:"type:jsonb"`
	CreatedAt   time.Time `gorm:"not null;index"`
}

func (auditLogModel) TableName() string {
	return "audit_logs"
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository crea una instancia del repositorio de auditoría respaldado por PostgreSQL.
func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &auditRepository{db: db}
}

// AutoMigrate crea o actualiza la tabla audit_logs y sus índices.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&auditLogModel{})
}

func (r *auditRepository) Create(ctx context.Context, logEntry *domain.AuditLog) error {
	return r.CreateTx(ctx, r.db, logEntry)
}

func (r *auditRepository) CreateTx(ctx context.Context, tx *gorm.DB, logEntry *domain.AuditLog) error {
	if tx == nil {
		tx = r.db
	}
	model := toAuditModel(logEntry)
	return tx.WithContext(ctx).Create(&model).Error
}

func (r *auditRepository) FindAll(ctx context.Context, filter domain.AuditFilter) (*domain.AuditListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&auditLogModel{})

	if filter.Module != "" {
		query = query.Where("module = ?", filter.Module)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.EntityID != "" {
		query = query.Where("entity_id = ?", filter.EntityID)
	}
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", *filter.To)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var models []auditLogModel
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]domain.AuditLog, 0, len(models))
	for _, m := range models {
		items = append(items, toAuditDomain(m))
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &domain.AuditListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func toAuditModel(l *domain.AuditLog) auditLogModel {
	return auditLogModel{
		ID:          l.ID,
		Module:      l.Module,
		Action:      l.Action,
		EntityType:  l.EntityType,
		EntityID:    l.EntityID,
		Description: l.Description,
		ActorID:     l.ActorID,
		ActorEmail:  l.ActorEmail,
		ActorRole:   l.ActorRole,
		IPAddress:   l.IPAddress,
		UserAgent:   l.UserAgent,
		OldValues:   nonEmptyBytes(l.OldValues),
		NewValues:   nonEmptyBytes(l.NewValues),
		Metadata:    nonEmptyBytes(l.Metadata),
		CreatedAt:   l.CreatedAt,
	}
}

func toAuditDomain(m auditLogModel) domain.AuditLog {
	return domain.AuditLog{
		ID:          m.ID,
		Module:      m.Module,
		Action:      m.Action,
		EntityType:  m.EntityType,
		EntityID:    m.EntityID,
		Description: m.Description,
		ActorID:     m.ActorID,
		ActorEmail:  m.ActorEmail,
		ActorRole:   m.ActorRole,
		IPAddress:   m.IPAddress,
		UserAgent:   m.UserAgent,
		OldValues:   json.RawMessage(m.OldValues),
		NewValues:   json.RawMessage(m.NewValues),
		Metadata:    json.RawMessage(m.Metadata),
		CreatedAt:   m.CreatedAt,
	}
}

func nonEmptyBytes(b json.RawMessage) []byte {
	if len(b) == 0 {
		return nil
	}
	return []byte(b)
}
