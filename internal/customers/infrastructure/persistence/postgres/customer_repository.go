package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-go/internal/customers/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type customerModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name           string    `gorm:"type:varchar(255);not null"`
	DocumentType   string    `gorm:"type:varchar(50);not null"`
	DocumentNumber string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Email          *string   `gorm:"type:varchar(255)"`
	Phone          *string   `gorm:"type:varchar(50)"`
	MobilePhone    *string   `gorm:"type:varchar(50)"`
	Address        *string   `gorm:"type:text"`
	City           *string   `gorm:"type:varchar(100)"`
	Department     *string   `gorm:"type:varchar(100)"`
	BusinessName   *string   `gorm:"type:varchar(255)"`
	TaxCategory    *string   `gorm:"type:varchar(100)"`
	Notes          *string   `gorm:"type:text"`
	Status         string    `gorm:"type:varchar(50);default:'active'"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (customerModel) TableName() string {
	return "customers"
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&customerModel{})
}

func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	model := toCustomerModel(customer)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrCustomerDocumentExists
		}
		return err
	}
	return nil
}

func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	var model customerModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, err
	}
	return toCustomerDomain(&model), nil
}

func (r *customerRepository) GetAll(ctx context.Context) ([]*domain.Customer, error) {
	var models []customerModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	customers := make([]*domain.Customer, len(models))
	for i, m := range models {
		customers[i] = toCustomerDomain(&m)
	}
	return customers, nil
}

func (r *customerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	model := toCustomerModel(customer)
	err := r.db.WithContext(ctx).Model(&customerModel{}).Where("id = ?", customer.ID).Updates(map[string]interface{}{
		"name":            model.Name,
		"document_type":   model.DocumentType,
		"document_number": model.DocumentNumber,
		"email":           model.Email,
		"phone":           model.Phone,
		"mobile_phone":    model.MobilePhone,
		"address":         model.Address,
		"city":            model.City,
		"department":      model.Department,
		"business_name":   model.BusinessName,
		"tax_category":    model.TaxCategory,
		"notes":           model.Notes,
		"status":          model.Status,
		"updated_at":      model.UpdatedAt,
	}).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrCustomerDocumentExists
		}
		return err
	}
	return nil
}

func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&customerModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrCustomerNotFound
	}
	return nil
}

func toCustomerDomain(m *customerModel) *domain.Customer {
	return &domain.Customer{
		ID:             m.ID,
		Name:           m.Name,
		DocumentType:   m.DocumentType,
		DocumentNumber: m.DocumentNumber,
		Email:          m.Email,
		Phone:          m.Phone,
		MobilePhone:    m.MobilePhone,
		Address:        m.Address,
		City:           m.City,
		Department:     m.Department,
		BusinessName:   m.BusinessName,
		TaxCategory:    m.TaxCategory,
		Notes:          m.Notes,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func toCustomerModel(c *domain.Customer) *customerModel {
	return &customerModel{
		ID:             c.ID,
		Name:           c.Name,
		DocumentType:   c.DocumentType,
		DocumentNumber: c.DocumentNumber,
		Email:          c.Email,
		Phone:          c.Phone,
		MobilePhone:    c.MobilePhone,
		Address:        c.Address,
		City:           c.City,
		Department:     c.Department,
		BusinessName:   c.BusinessName,
		TaxCategory:    c.TaxCategory,
		Notes:          c.Notes,
		Status:         c.Status,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505"))
}
