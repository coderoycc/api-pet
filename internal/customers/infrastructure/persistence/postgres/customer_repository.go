package postgres

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"api-go/internal/customers/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// allowedSortColumns es la whitelist de columnas por las que se puede ordenar.
var allowedSortColumns = map[string]string{
	"name":           "name",
	"documentnumber": "document_number",
	"city":           "city",
	"status":         "status",
	"createdat":      "created_at",
	"updatedat":      "updated_at",
	"created_at":     "created_at",
	"updated_at":     "updated_at",
}

type customerModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name           string    `gorm:"type:varchar(255);not null;index:idx_customers_name"`
	DocumentType   string    `gorm:"type:varchar(50);not null"`
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

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&customerModel{})
}

// Create inserta un nuevo cliente en la base de datos.
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

// GetByID obtiene un cliente por su UUID.
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

// GetAll retorna la lista paginada y filtrada de clientes.
func (r *customerRepository) GetAll(
	ctx context.Context,
	page, limit int,
	filters *domain.CustomerFilters,
	sortBy string,
	descending bool,
) (*domain.PaginatedCustomerResult, error) {
	var models []customerModel
	var total int64

	query := r.db.WithContext(ctx).Model(&customerModel{})

	if filters != nil {
		if filters.Search != "" {
			search := "%" + filters.Search + "%"
			query = query.Where(
				"name ILIKE ? OR document_number ILIKE ? OR business_name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR mobile_phone ILIKE ?",
				search, search, search, search, search, search,
			)
		}
		if len(filters.DocumentTypes) > 0 {
			query = query.Where("document_type IN ?", filters.DocumentTypes)
		}
		if len(filters.Statuses) > 0 {
			query = query.Where("status IN ?", filters.Statuses)
		}
		if filters.City != "" {
			query = query.Where("city ILIKE ?", filters.City)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Ordenamiento seguro usando whitelist
	orderColumn := "created_at"
	if col, ok := allowedSortColumns[strings.ToLower(sortBy)]; ok {
		orderColumn = col
	}
	direction := "ASC"
	if descending {
		direction = "DESC"
	}
	query = query.Order(orderColumn + " " + direction)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	customers := make([]*domain.Customer, len(models))
	for i, m := range models {
		customers[i] = toCustomerDomain(&m)
	}

	return &domain.PaginatedCustomerResult{
		Data:       customers,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

// Search realiza búsqueda rápida para selectores/autocomplete.
func (r *customerRepository) Search(ctx context.Context, query string, limit int, status string) ([]*domain.Customer, error) {
	var models []customerModel
	search := "%" + query + "%"

	q := r.db.WithContext(ctx).
		Where(
			"name ILIKE ? OR document_number ILIKE ? OR business_name ILIKE ? OR email ILIKE ? OR mobile_phone ILIKE ?",
			search, search, search, search, search,
		).
		Order("name ASC").
		Limit(limit)

	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	customers := make([]*domain.Customer, len(models))
	for i, m := range models {
		customers[i] = toCustomerDomain(&m)
	}
	return customers, nil
}

// GetDistinctCities retorna la lista de ciudades únicas no nulas ordenada alfabéticamente.
func (r *customerRepository) GetDistinctCities(ctx context.Context) ([]string, error) {
	var cities []string
	err := r.db.WithContext(ctx).
		Model(&customerModel{}).
		Where("city IS NOT NULL AND city != ''").
		Distinct("city").
		Order("city ASC").
		Pluck("city", &cities).Error
	if err != nil {
		return nil, err
	}
	if cities == nil {
		cities = []string{}
	}
	return cities, nil
}

// Update actualiza los datos de un cliente existente.
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

// Delete elimina físicamente un cliente de la base de datos.
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

// CheckDocumentConflict verifica si ya existe un cliente con el mismo tipo y número de documento.
// excludeID permite excluir un registro específico (útil en actualizaciones).
func (r *customerRepository) CheckDocumentConflict(ctx context.Context, docType, docNumber string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&customerModel{}).
		Where("document_type = ? AND document_number = ?", docType, docNumber)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasAssociatedRecords verifica si el cliente tiene registros vinculados en otras tablas.
// Actualmente revisa si existe como customer_id en inventory_transactions.
func (r *customerRepository) HasAssociatedRecords(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	// Verificar si la tabla inventory_transactions tiene registros con este customer_id
	err := r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'inventory_transactions'").
		Scan(&count).Error
	if err != nil {
		return false, err
	}

	// Si la tabla no existe aún, no hay registros vinculados
	if count == 0 {
		return false, nil
	}

	// Verificar registros vinculados en inventory_transactions
	var linked int64
	err = r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM inventory_transactions WHERE customer_id = ?", id).
		Scan(&linked).Error
	if err != nil {
		return false, err
	}
	return linked > 0, nil
}

// --- Helpers ---

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
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(err.Error(), "duplicate key value") ||
		strings.Contains(err.Error(), "23505"))
}
