package application

import (
	"context"
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

func (s *CustomerService) CreateCustomer(ctx context.Context, dto CustomerCreateDTO) (*domain.Customer, error) {
	status := dto.Status
	if status == "" {
		status = "active" // Default status
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

	err := s.repo.Create(ctx, customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CustomerService) GetAllCustomers(ctx context.Context) ([]*domain.Customer, error) {
	return s.repo.GetAll(ctx)
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, id uuid.UUID, dto CustomerUpdateDTO) (*domain.Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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

	err = s.repo.Update(ctx, customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
