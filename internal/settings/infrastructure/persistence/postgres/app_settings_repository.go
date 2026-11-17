package postgres

import (
	"context"
	"time"

	"api-go/internal/settings/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type appSettingsModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	CurrencySymbol   string    `gorm:"type:varchar(10);not null;default:'Bs'"`
	CurrencyCode     string    `gorm:"type:varchar(10);not null;default:'BOB'"`
	CurrencyPosition string    `gorm:"type:varchar(10);not null;default:'suffix'"`
	DecimalPlaces    int       `gorm:"not null;default:2"`
	CompanyName      string    `gorm:"type:varchar(255);not null;default:'Pet Shop & Veterinaria'"`
	CompanyTaxID     string    `gorm:"type:varchar(50)"`
	CompanyPhone     string    `gorm:"type:varchar(50)"`
	CompanyAddress   string    `gorm:"type:varchar(255)"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (appSettingsModel) TableName() string {
	return "app_settings"
}

type appSettingsRepository struct {
	db *gorm.DB
}

func NewAppSettingsRepository(db *gorm.DB) domain.AppSettingsRepository {
	return &appSettingsRepository{db: db}
}

func (r *appSettingsRepository) Get(ctx context.Context) (*domain.AppSettings, error) {
	var model appSettingsModel
	err := r.db.WithContext(ctx).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Si no existe, creamos una por defecto
			now := time.Now().UTC()
			model = appSettingsModel{
				ID:               uuid.New(),
				CurrencySymbol:   "Bs",
				CurrencyCode:     "BOB",
				CurrencyPosition: "suffix",
				DecimalPlaces:    2,
				CompanyName:      "Pet Shop & Veterinaria",
				CompanyTaxID:     "",
				CompanyPhone:     "",
				CompanyAddress:   "",
				UpdatedAt:        now,
			}
			if createErr := r.db.WithContext(ctx).Create(&model).Error; createErr != nil {
				return nil, createErr
			}
		} else {
			return nil, err
		}
	}

	return toDomainAppSettings(&model), nil
}

func (r *appSettingsRepository) Update(ctx context.Context, s *domain.AppSettings) error {
	model := appSettingsModel{
		ID:               s.ID,
		CurrencySymbol:   s.CurrencySymbol,
		CurrencyCode:     s.CurrencyCode,
		CurrencyPosition: s.CurrencyPosition,
		DecimalPlaces:    s.DecimalPlaces,
		CompanyName:      s.CompanyName,
		CompanyTaxID:     s.CompanyTaxID,
		CompanyPhone:     s.CompanyPhone,
		CompanyAddress:   s.CompanyAddress,
		UpdatedAt:        s.UpdatedAt,
	}
	return r.db.WithContext(ctx).Save(&model).Error
}

func toDomainAppSettings(m *appSettingsModel) *domain.AppSettings {
	return &domain.AppSettings{
		ID:               m.ID,
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
