package postgres

import (
	"time"

	"api-go/internal/customers/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type customerModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name           string    `gorm:"type:varchar(255);not null;index:idx_customers_name"`
	DocumentType   string    `gorm:"type:varchar(50);not null;uniqueIndex:uq_customers_doc"`
	DocumentNumber string    `gorm:"type:varchar(50);not null;uniqueIndex:uq_customers_doc"`
	Email          *string   `gorm:"type:varchar(255)"`
	Phone          *string   `gorm:"type:varchar(50)"`
	MobilePhone    *string   `gorm:"type:varchar(50)"`
	Address        *string   `gorm:"type:text"`
	City           *string   `gorm:"type:varchar(100)"`
	Department     *string   `gorm:"type:varchar(100)"`
	BusinessName   *string   `gorm:"type:varchar(255)"`
	TaxCategory    *string   `gorm:"type:varchar(100)"`
	Notes          *string   `gorm:"type:text"`
	Status         string    `gorm:"type:varchar(50);default:'active';index:idx_customers_status"`
	CreatedAt      time.Time `gorm:"not null;index:idx_customers_created_at"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (customerModel) TableName() string {
	return "customers"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&customerModel{})
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
