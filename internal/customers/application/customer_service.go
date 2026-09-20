package application

import (
	"context"
	"fmt"
	"math"
	"net/mail"
	"time"

	"api-go/internal/customers/domain"

	"github.com/google/uuid"
)

// validDocumentTypes define los tipos de documento aceptados.
var validDocumentTypes = map[string]bool{
	"nit":      true,
	"ci":       true,
	"passport": true,
	"other":    true,
}

type CustomerService struct {
	repo domain.CustomerRepository
}

func NewCustomerService(repo domain.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

// validateDocumentType verifica que el tipo de documento sea válido.
func validateDocumentType(docType string) error {
	if !validDocumentTypes[docType] {
		return domain.ErrInvalidCustomerDocumentType
	}
	return nil
}

// validateEmail verifica el formato del email si se proporciona.
func validateEmail(email *string) error {
	if email == nil || *email == "" {
		return nil
	}
	if _, err := mail.ParseAddress(*email); err != nil {
		return fmt.Errorf("invalid email format: %s", *email)
	}
	return nil
}

// CreateCustomer crea un nuevo cliente con validaciones completas.
func (s *CustomerService) CreateCustomer(ctx context.Context, dto CustomerCreateDTO) (*domain.Customer, error) {
	if dto.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if dto.DocumentType == "" {
		return nil, fmt.Errorf("documentType is required")
	}
	if dto.DocumentNumber == "" {
		return nil, fmt.Errorf("documentNumber is required")
	}

	if err := validateDocumentType(dto.DocumentType); err != nil {
		return nil, err
	}
	if err := validateEmail(dto.Email); err != nil {
		return nil, err
	}

	// Verificar unicidad del documento
	conflict, err := s.repo.CheckDocumentConflict(ctx, dto.DocumentType, dto.DocumentNumber, nil)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, domain.ErrCustomerDocumentExists
	}

	status := dto.Status
	if status == "" {
		status = "active"
	}

	customer := &domain.Customer{
		ID:             uuid.New(),
		Name:           dto.Name,
		DocumentType:   dto.DocumentType,
		DocumentNumber: dto.DocumentNumber,
		Email:          dto.Email,
		Phone:          dto.Phone,
		MobilePhone:    dto.MobilePhone,
		Address:        dto.Address,
		City:           dto.City,
		Department:     dto.Department,
		BusinessName:   dto.BusinessName,
		TaxCategory:    dto.TaxCategory,
		Notes:          dto.Notes,
		Status:         status,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomerByID obtiene un cliente por su ID.
func (s *CustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

// GetCustomers retorna la lista paginada y filtrada de clientes.
func (s *CustomerService) GetCustomers(
	ctx context.Context,
	page, limit int,
	filters *domain.CustomerFilters,
	sortBy string,
	descending bool,
) (*CustomerListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if sortBy == "" {
		sortBy = "created_at"
	}

	result, err := s.repo.GetAll(ctx, page, limit, filters, sortBy, descending)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(limit)))
	if totalPages < 1 && result.Total == 0 {
		totalPages = 0
	}

	return &CustomerListResponse{
		Data: result.Data,
		Pagination: CustomerPaginationMeta{
			CurrentPage: page,
			PageSize:    limit,
			TotalItems:  result.Total,
			TotalPages:  totalPages,
		},
		Total: result.Total,
	}, nil
}

// SearchCustomers retorna una lista ligera de clientes para selectores/autocomplete.
func (s *CustomerService) SearchCustomers(ctx context.Context, query string, limit int, status string) ([]CustomerSelectorDTO, error) {
	if len(query) < 2 {
		return []CustomerSelectorDTO{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if status == "" {
		status = "active"
	}

	customers, err := s.repo.Search(ctx, query, limit, status)
	if err != nil {
		return nil, err
	}

	result := make([]CustomerSelectorDTO, len(customers))
	for i, c := range customers {
		result[i] = CustomerSelectorDTO{
			ID:             c.ID,
			Name:           c.Name,
			DocumentType:   c.DocumentType,
			DocumentNumber: c.DocumentNumber,
			BusinessName:   c.BusinessName,
			Email:          c.Email,
			MobilePhone:    c.MobilePhone,
			Status:         c.Status,
		}
	}
	return result, nil
}

// GetDistinctCities retorna la lista de ciudades únicas registradas.
func (s *CustomerService) GetDistinctCities(ctx context.Context) ([]string, error) {
	return s.repo.GetDistinctCities(ctx)
}

// UpdateCustomer actualiza los datos de un cliente con validaciones completas.
func (s *CustomerService) UpdateCustomer(ctx context.Context, id uuid.UUID, dto CustomerUpdateDTO) (*domain.Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := validateDocumentType(dto.DocumentType); err != nil {
		return nil, err
	}
	if err := validateEmail(dto.Email); err != nil {
		return nil, err
	}

	// Verificar unicidad del documento excluyendo al propio cliente
	conflict, err := s.repo.CheckDocumentConflict(ctx, dto.DocumentType, dto.DocumentNumber, &id)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, domain.ErrCustomerDocumentExists
	}

	customer.Name = dto.Name
	customer.DocumentType = dto.DocumentType
	customer.DocumentNumber = dto.DocumentNumber
	customer.Email = dto.Email
	customer.Phone = dto.Phone
	customer.MobilePhone = dto.MobilePhone
	customer.Address = dto.Address
	customer.City = dto.City
	customer.Department = dto.Department
	customer.BusinessName = dto.BusinessName
	customer.TaxCategory = dto.TaxCategory
	customer.Notes = dto.Notes

	if dto.Status != "" {
		customer.Status = dto.Status
	}
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// DeleteCustomer elimina un cliente, verificando primero la integridad referencial.
func (s *CustomerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	hasRecords, err := s.repo.HasAssociatedRecords(ctx, id)
	if err != nil {
		return err
	}
	if hasRecords {
		return domain.ErrCustomerHasAssociatedRecords
	}
	return s.repo.Delete(ctx, id)
}
