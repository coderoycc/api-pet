package application

import (
	"context"
	"time"

	"api-go/internal/settings/domain"
)

type AppSettingsService interface {
	GetSettings(ctx context.Context) (*AppSettingsResponseDTO, error)
	UpdateSettings(ctx context.Context, dto UpdateAppSettingsDTO) (*AppSettingsResponseDTO, error)
}

type appSettingsService struct {
	repo domain.AppSettingsRepository
}

func NewAppSettingsService(repo domain.AppSettingsRepository) AppSettingsService {
	return &appSettingsService{repo: repo}
}

func (s *appSettingsService) GetSettings(ctx context.Context) (*AppSettingsResponseDTO, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return toAppSettingsResponseDTO(settings), nil
}

func (s *appSettingsService) UpdateSettings(ctx context.Context, dto UpdateAppSettingsDTO) (*AppSettingsResponseDTO, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	if dto.CurrencySymbol != "" {
		settings.CurrencySymbol = dto.CurrencySymbol
	}
	if dto.CurrencyCode != "" {
		settings.CurrencyCode = dto.CurrencyCode
	}
	if dto.CurrencyPosition != "" {
		settings.CurrencyPosition = dto.CurrencyPosition
	}
	if dto.DecimalPlaces >= 0 && dto.DecimalPlaces <= 4 {
		settings.DecimalPlaces = dto.DecimalPlaces
	}
	if dto.CompanyName != "" {
		settings.CompanyName = dto.CompanyName
	}
	settings.CompanyTaxID = dto.CompanyTaxID
	settings.CompanyPhone = dto.CompanyPhone
	settings.CompanyAddress = dto.CompanyAddress
	settings.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return toAppSettingsResponseDTO(settings), nil
}

func toAppSettingsResponseDTO(m *domain.AppSettings) *AppSettingsResponseDTO {
	return &AppSettingsResponseDTO{
		ID:               m.ID.String(),
		CurrencySymbol:   m.CurrencySymbol,
		CurrencyCode:     m.CurrencyCode,
		CurrencyPosition: m.CurrencyPosition,
		DecimalPlaces:    m.DecimalPlaces,
		CompanyName:      m.CompanyName,
		CompanyTaxID:     m.CompanyTaxID,
		CompanyPhone:     m.CompanyPhone,
		CompanyAddress:   m.CompanyAddress,
		UpdatedAt:        m.UpdatedAt,
	}
}
