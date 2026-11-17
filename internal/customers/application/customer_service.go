package application

import (
	"context"
	"math"
	"time"

	"api-go/internal/customers/domain"

	"github.com/google/uuid"
)

type CustomerService struct {
	repo domain.CustomerRepository
}

func NewCustomerService(repo domain.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

// CreateCustomer crea un nuevo cliente con validaciones y normalización completas.
func (s *CustomerService) CreateCustomer(ctx context.Context, dto CustomerCreateDTO) (*domain.Customer, error) {
	input := customerInput{
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
		Status:         dto.Status,
	}

	if err := normalizeAndValidateCustomer(&input, domain.StatusActive); err != nil {
		return nil, err
	}

	conflict, err := s.repo.CheckDocumentConflict(ctx, input.DocumentType, input.DocumentNumber, nil)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, domain.ErrCustomerDocumentExists
	}

	now := time.Now()
	customer := &domain.Customer{
		ID:             uuid.New(),
		Name:           input.Name,
		DocumentType:   input.DocumentType,
		DocumentNumber: input.DocumentNumber,
		Email:          input.Email,
		Phone:          input.Phone,
		MobilePhone:    input.MobilePhone,
		Address:        input.Address,
		City:           input.City,
		Department:     input.Department,
		BusinessName:   input.BusinessName,
		TaxCategory:    input.TaxCategory,
		Notes:          input.Notes,
		Status:         input.Status,
		CreatedAt:      now,
		UpdatedAt:      now,
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
	if result.Total == 0 {
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
		status = domain.StatusActive
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

// UpdateCustomer actualiza los datos de un cliente con validaciones y normalización completas.
func (s *CustomerService) UpdateCustomer(ctx context.Context, id uuid.UUID, dto CustomerUpdateDTO) (*domain.Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	input := customerInput{
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
		Status:         dto.Status,
	}

	if err := normalizeAndValidateCustomer(&input, customer.Status); err != nil {
		return nil, err
	}

	conflict, err := s.repo.CheckDocumentConflict(ctx, input.DocumentType, input.DocumentNumber, &id)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, domain.ErrCustomerDocumentExists
	}

	customer.Name = input.Name
	customer.DocumentType = input.DocumentType
	customer.DocumentNumber = input.DocumentNumber
	customer.Email = input.Email
	customer.Phone = input.Phone
	customer.MobilePhone = input.MobilePhone
	customer.Address = input.Address
	customer.City = input.City
	customer.Department = input.Department
	customer.BusinessName = input.BusinessName
	customer.TaxCategory = input.TaxCategory
	customer.Notes = input.Notes
	customer.Status = input.Status
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// DeleteCustomer elimina un cliente, verificando primero la integridad referencial.
func (s *CustomerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	hasRecords, err := s.repo.HasAssociatedRecords(ctx, id)
	if err != nil {
		return err
	}
	if hasRecords {
		return domain.ErrCustomerHasAssociatedRecords
	}

	return s.repo.Delete(ctx, id)
}
