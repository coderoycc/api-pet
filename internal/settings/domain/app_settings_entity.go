package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AppSettings struct {
	ID               uuid.UUID `json:"id"`
	CurrencySymbol   string    `json:"currencySymbol"`   // ej: "Bs", "$", "USD"
	CurrencyCode     string    `json:"currencyCode"`     // ej: "BOB", "USD"
	CurrencyPosition string    `json:"currencyPosition"` // "prefix" ($ 10) o "suffix" (10 Bs)
	DecimalPlaces    int       `json:"decimalPlaces"`    // 2 por defecto
	CompanyName      string    `json:"companyName"`      // Nombre comercial
	CompanyTaxID     string    `json:"companyTaxId"`     // NIT / RUT / CUIT
	CompanyPhone     string    `json:"companyPhone"`     // Teléfono de contacto
	CompanyAddress   string    `json:"companyAddress"`   // Dirección
	UpdatedAt        time.Time `json:"updatedAt"`
}

type AppSettingsRepository interface {
	Get(ctx context.Context) (*AppSettings, error)
	Update(ctx context.Context, settings *AppSettings) error
}
