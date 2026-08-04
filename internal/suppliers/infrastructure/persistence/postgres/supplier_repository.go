package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-go/internal/suppliers/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type supplierModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"type:varchar(255);not null"`
	TaxID        string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	ContactEmail string    `gorm:"type:varchar(255)"`
	ContactPhone string    `gorm:"type:varchar(50)"`
	Address      string    `gorm:"type:text"`
	Status       bool      `gorm:"type:boolean;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (supplierModel) TableName() string {
	return "suppliers"
}

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) domain.SupplierRepository {
	return &supplierRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&supplierModel{})
}

func (r *supplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	model := toSupplierModel(supplier)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrSupplierTaxIdExists
		}
		return err
	}
	return nil
}

func (r *supplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	var model supplierModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}
	return toSupplierDomain(&model), nil
}

func (r *supplierRepository) GetAll(ctx context.Context) ([]*domain.Supplier, error) {
	var models []supplierModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	suppliers := make([]*domain.Supplier, len(models))
	for i, m := range models {
		suppliers[i] = toSupplierDomain(&m)
	}
	return suppliers, nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	model := toSupplierModel(supplier)
	err := r.db.WithContext(ctx).Model(&supplierModel{}).Where("id = ?", supplier.ID).Updates(map[string]interface{}{
		"name":          model.Name,
		"tax_id":        model.TaxID,
		"contact_email": model.ContactEmail,
		"contact_phone": model.ContactPhone,
		"address":       model.Address,
		"status":        model.Status,
		"updated_at":    model.UpdatedAt,
	}).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrSupplierTaxIdExists
		}
		return err
	}
	return nil
}

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&supplierModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSupplierNotFound
	}
	return nil
}

func toSupplierDomain(m *supplierModel) *domain.Supplier {
	return &domain.Supplier{
		ID:           m.ID,
		Name:         m.Name,
		TaxID:        m.TaxID,
		ContactEmail: m.ContactEmail,
		ContactPhone: m.ContactPhone,
		Address:      m.Address,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toSupplierModel(s *domain.Supplier) *supplierModel {
	return &supplierModel{
		ID:           s.ID,
		Name:         s.Name,
		TaxID:        s.TaxID,
		ContactEmail: s.ContactEmail,
		ContactPhone: s.ContactPhone,
		Address:      s.Address,
		Status:       s.Status,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505"))
}
