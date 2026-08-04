package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-go/internal/settings/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type unitOfMeasureModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Detail       string    `gorm:"type:varchar(255);not null"`
	Abbreviation string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Status       string    `gorm:"type:varchar(50);default:'active'"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (unitOfMeasureModel) TableName() string {
	return "unit_of_measures"
}

type unitsRepository struct {
	db *gorm.DB
}

func NewUnitOfMeasureRepository(db *gorm.DB) domain.UnitOfMeasureRepository {
	return &unitsRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&unitOfMeasureModel{},
		&warehouseModel{},
	)
}

func (r *unitsRepository) Create(ctx context.Context, uom *domain.UnitOfMeasure) error {
	model := toUnitOfMeasureModel(uom)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrUnitAbbrevExists
		}
		return err
	}
	return nil
}

func (r *unitsRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error) {
	var model unitOfMeasureModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, err
	}
	return toUnitOfMeasureDomain(&model), nil
}

func (r *unitsRepository) GetAll(ctx context.Context) ([]*domain.UnitOfMeasure, error) {
	var models []unitOfMeasureModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	uoms := make([]*domain.UnitOfMeasure, len(models))
	for i, m := range models {
		uoms[i] = toUnitOfMeasureDomain(&m)
	}
	return uoms, nil
}

func (r *unitsRepository) Update(ctx context.Context, uom *domain.UnitOfMeasure) error {
	model := toUnitOfMeasureModel(uom)
	err := r.db.WithContext(ctx).Model(&unitOfMeasureModel{}).Where("id = ?", uom.ID).Updates(map[string]interface{}{
		"detail":       model.Detail,
		"abbreviation": model.Abbreviation,
		"status":       model.Status,
		"updated_at":   model.UpdatedAt,
	}).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrUnitAbbrevExists
		}
		return err
	}
	return nil
}

func (r *unitsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&unitOfMeasureModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUnitNotFound
	}
	return nil
}

func toUnitOfMeasureDomain(m *unitOfMeasureModel) *domain.UnitOfMeasure {
	return &domain.UnitOfMeasure{
		ID:           m.ID,
		Detail:       m.Detail,
		Abbreviation: m.Abbreviation,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toUnitOfMeasureModel(u *domain.UnitOfMeasure) *unitOfMeasureModel {
	return &unitOfMeasureModel{
		ID:           u.ID,
		Detail:       u.Detail,
		Abbreviation: u.Abbreviation,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505"))
}
