package postgres

import (
	"context"
	"errors"
	"time"

	"api-go/internal/settings/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type warehouseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Address   string    `gorm:"type:varchar(255);not null"`
	Location  string    `gorm:"type:varchar(255);not null"`
	Type      string    `gorm:"type:varchar(50);not null"`
	Status    bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (warehouseModel) TableName() string {
	return "warehouses"
}

type warehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) domain.WarehouseRepository {
	return &warehouseRepository{db: db}
}



func (r *warehouseRepository) Create(ctx context.Context, w *domain.Warehouse) error {
	model := toWarehouseModel(w)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrWarehouseNameExists
		}
		return err
	}
	return nil
}

func (r *warehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	var model warehouseModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrWarehouseNotFound
		}
		return nil, err
	}
	return toWarehouseDomain(&model), nil
}

func (r *warehouseRepository) GetAll(ctx context.Context) ([]*domain.Warehouse, error) {
	var models []warehouseModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	warehouses := make([]*domain.Warehouse, len(models))
	for i, m := range models {
		warehouses[i] = toWarehouseDomain(&m)
	}
	return warehouses, nil
}

func (r *warehouseRepository) Update(ctx context.Context, w *domain.Warehouse) error {
	model := toWarehouseModel(w)
	err := r.db.WithContext(ctx).Model(&warehouseModel{}).Where("id = ?", w.ID).Updates(map[string]interface{}{
		"name":       model.Name,
		"address":    model.Address,
		"location":   model.Location,
		"type":       model.Type,
		"status":     model.Status,
		"updated_at": model.UpdatedAt,
	}).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrWarehouseNameExists
		}
		return err
	}
	return nil
}

func (r *warehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&warehouseModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrWarehouseNotFound
	}
	return nil
}

func toWarehouseDomain(m *warehouseModel) *domain.Warehouse {
	return &domain.Warehouse{
		ID:        m.ID,
		Name:      m.Name,
		Address:   m.Address,
		Location:  m.Location,
		Type:      m.Type,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toWarehouseModel(w *domain.Warehouse) *warehouseModel {
	return &warehouseModel{
		ID:        w.ID,
		Name:      w.Name,
		Address:   w.Address,
		Location:  w.Location,
		Type:      w.Type,
		Status:    w.Status,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}


