package application

import (
	"context"
	"time"

	"api-go/internal/settings/domain"

	"github.com/google/uuid"
)

type WarehouseService struct {
	repo domain.WarehouseRepository
}

func NewWarehouseService(repo domain.WarehouseRepository) *WarehouseService {
	return &WarehouseService{repo: repo}
}

func (s *WarehouseService) Create(ctx context.Context, dto CreateWarehouseDTO) (*domain.Warehouse, error) {
	w := &domain.Warehouse{
		ID:        uuid.New(),
		Name:      dto.Name,
		Address:   dto.Address,
		Location:  dto.Location,
		Type:      dto.Type,
		Status:    dto.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}

	return w, nil
}

func (s *WarehouseService) GetByID(ctx context.Context, idStr string) (*domain.Warehouse, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *WarehouseService) GetAll(ctx context.Context) ([]*domain.Warehouse, error) {
	return s.repo.GetAll(ctx)
}

func (s *WarehouseService) Update(ctx context.Context, idStr string, dto UpdateWarehouseDTO) (*domain.Warehouse, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	w.Name = dto.Name
	w.Address = dto.Address
	w.Location = dto.Location
	w.Type = dto.Type
	w.Status = dto.Status
	w.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}

	return w, nil
}

func (s *WarehouseService) Delete(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func MapWarehouseToResponse(w *domain.Warehouse) WarehouseResponseDTO {
	return WarehouseResponseDTO{
		ID:        w.ID.String(),
		Name:      w.Name,
		Address:   w.Address,
		Location:  w.Location,
		Type:      w.Type,
		Status:    w.Status,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func MapWarehouseToResponseList(warehouses []*domain.Warehouse) []WarehouseResponseDTO {
	var list []WarehouseResponseDTO
	for _, w := range warehouses {
		list = append(list, MapWarehouseToResponse(w))
	}
	if list == nil {
		list = make([]WarehouseResponseDTO, 0)
	}
	return list
}
