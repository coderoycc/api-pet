package application

import (
	"context"
	"time"

	"api-go/internal/suppliers/domain"

	"github.com/google/uuid"
)

type SupplierService interface {
	Create(ctx context.Context, dto CreateSupplierDTO) (*SupplierResponseDTO, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SupplierResponseDTO, error)
	GetAll(ctx context.Context) ([]*SupplierResponseDTO, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateSupplierDTO) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type supplierService struct {
	repo domain.SupplierRepository
}

func NewSupplierService(repo domain.SupplierRepository) SupplierService {
	return &supplierService{repo: repo}
}

func (s *supplierService) Create(ctx context.Context, dto CreateSupplierDTO) (*SupplierResponseDTO, error) {
	supplier := &domain.Supplier{
		ID:           uuid.New(),
		Name:         dto.Name,
		TaxID:        dto.TaxID,
		ContactEmail: dto.ContactEmail,
		ContactPhone: dto.ContactPhone,
		Address:      dto.Address,
		Status:       dto.Status,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, err
	}

	return toSupplierResponseDTO(supplier), nil
}

func (s *supplierService) GetByID(ctx context.Context, id uuid.UUID) (*SupplierResponseDTO, error) {
	supplier, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toSupplierResponseDTO(supplier), nil
}

func (s *supplierService) GetAll(ctx context.Context) ([]*SupplierResponseDTO, error) {
	suppliers, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []*SupplierResponseDTO
	for _, supplier := range suppliers {
		dtos = append(dtos, toSupplierResponseDTO(supplier))
	}
	return dtos, nil
}

func (s *supplierService) Update(ctx context.Context, id uuid.UUID, dto UpdateSupplierDTO) error {
	supplier, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	supplier.Name = dto.Name
	supplier.TaxID = dto.TaxID
	supplier.ContactEmail = dto.ContactEmail
	supplier.ContactPhone = dto.ContactPhone
	supplier.Address = dto.Address
	supplier.Status = dto.Status
	supplier.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, supplier)
}

func (s *supplierService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func toSupplierResponseDTO(supplier *domain.Supplier) *SupplierResponseDTO {
	return &SupplierResponseDTO{
		ID:           supplier.ID.String(),
		Name:         supplier.Name,
		TaxID:        supplier.TaxID,
		ContactEmail: supplier.ContactEmail,
		ContactPhone: supplier.ContactPhone,
		Address:      supplier.Address,
		Status:       supplier.Status,
		CreatedAt:    supplier.CreatedAt,
		UpdatedAt:    supplier.UpdatedAt,
	}
}
