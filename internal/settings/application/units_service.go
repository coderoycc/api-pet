package application

import (
	"context"
	"strings"
	"time"

	"api-go/internal/settings/domain"

	"github.com/google/uuid"
)

type UnitOfMeasureService struct {
	repo domain.UnitOfMeasureRepository
}

func NewUnitOfMeasureService(repo domain.UnitOfMeasureRepository) *UnitOfMeasureService {
	return &UnitOfMeasureService{repo: repo}
}

func (s *UnitOfMeasureService) Create(ctx context.Context, dto CreateUnitOfMeasureDTO) (*domain.UnitOfMeasure, error) {
	uom := &domain.UnitOfMeasure{
		ID:           uuid.New(),
		Detail:       dto.Detail,
		Abbreviation: strings.ToUpper(dto.Abbreviation),
		Status:       dto.Status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, uom); err != nil {
		return nil, err
	}

	return uom, nil
}

func (s *UnitOfMeasureService) GetByID(ctx context.Context, idStr string) (*domain.UnitOfMeasure, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UnitOfMeasureService) GetAll(ctx context.Context) ([]*domain.UnitOfMeasure, error) {
	return s.repo.GetAll(ctx)
}

func (s *UnitOfMeasureService) Update(ctx context.Context, idStr string, dto UpdateUnitOfMeasureDTO) (*domain.UnitOfMeasure, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	uom, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	uom.Detail = dto.Detail
	uom.Abbreviation = strings.ToUpper(dto.Abbreviation)
	uom.Status = dto.Status
	uom.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, uom); err != nil {
		return nil, err
	}

	return uom, nil
}

func (s *UnitOfMeasureService) Delete(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func MapUnitOfMeasureToResponse(u *domain.UnitOfMeasure) UnitOfMeasureResponseDTO {
	return UnitOfMeasureResponseDTO{
		ID:           u.ID.String(),
		Detail:       u.Detail,
		Abbreviation: u.Abbreviation,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func MapUnitOfMeasureToResponseList(units []*domain.UnitOfMeasure) []UnitOfMeasureResponseDTO {
	var list []UnitOfMeasureResponseDTO
	for _, u := range units {
		list = append(list, MapUnitOfMeasureToResponse(u))
	}
	// Return empty array instead of null for JSON
	if list == nil {
		list = make([]UnitOfMeasureResponseDTO, 0)
	}
	return list
}
