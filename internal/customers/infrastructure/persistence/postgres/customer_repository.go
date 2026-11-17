package postgres

import (
	"context"
	"errors"
	"math"
	"strings"

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

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{db: db}
}

// Create inserta un nuevo cliente en la base de datos.
func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	model := toCustomerModel(customer)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
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

	query := applyCustomerFilters(r.db.WithContext(ctx).Model(&customerModel{}), filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	orderClause := buildOrderClause(sortBy, descending)
	offset := (page - 1) * limit
	if err := query.Order(orderClause).Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedCustomerResult{
		Data:       mapCustomersToDomain(models),
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
			"(name ILIKE ? OR document_number ILIKE ? OR business_name ILIKE ? OR email ILIKE ? OR mobile_phone ILIKE ?)",
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

	return mapCustomersToDomain(models), nil
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
func (r *customerRepository) HasAssociatedRecords(ctx context.Context, id uuid.UUID) (bool, error) {
	tables := []string{"inventory_logs", "sales", "inventory_transactions"}
	for _, table := range tables {
		linked, err := r.countCustomerLinksInTable(ctx, table, id)
		if err != nil {
			return false, err
		}
		if linked > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (r *customerRepository) countCustomerLinksInTable(ctx context.Context, tableName string, customerID uuid.UUID) (int64, error) {
	var colExists int64
	err := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = 'customer_id'",
		tableName,
	).Scan(&colExists).Error
	if err != nil || colExists == 0 {
		return 0, err
	}

	var linked int64
	err = r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM "+tableName+" WHERE customer_id = ?", customerID).
		Scan(&linked).Error
	return linked, err
}

func applyCustomerFilters(query *gorm.DB, filters *domain.CustomerFilters) *gorm.DB {
	if filters == nil {
		return query
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where(
			"(name ILIKE ? OR document_number ILIKE ? OR business_name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR mobile_phone ILIKE ?)",
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
	return query
}

func buildOrderClause(sortBy string, descending bool) string {
	orderColumn := "created_at"
	if col, ok := allowedSortColumns[strings.ToLower(sortBy)]; ok {
		orderColumn = col
	}
	direction := "ASC"
	if descending {
		direction = "DESC"
	}
	return orderColumn + " " + direction
}

func mapCustomersToDomain(models []customerModel) []*domain.Customer {
	customers := make([]*domain.Customer, len(models))
	for i := range models {
		customers[i] = toCustomerDomain(&models[i])
	}
	return customers
}

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(err.Error(), "duplicate key value") ||
		strings.Contains(err.Error(), "23505"))
}
