package application

import "time"

type AppSettingsResponseDTO struct {
	ID               string    `json:"id"`
	CurrencySymbol   string    `json:"currencySymbol"`
	CurrencyCode     string    `json:"currencyCode"`
	CurrencyPosition string    `json:"currencyPosition"`
	DecimalPlaces    int       `json:"decimalPlaces"`
	CompanyName      string    `json:"companyName"`
	CompanyTaxID     string    `json:"companyTaxId"`
	CompanyPhone     string    `json:"companyPhone"`
	CompanyAddress   string    `json:"companyAddress"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type UpdateAppSettingsDTO struct {
	CurrencySymbol   string `json:"currencySymbol"`
	CurrencyCode     string `json:"currencyCode"`
	CurrencyPosition string `json:"currencyPosition"`
	DecimalPlaces    int    `json:"decimalPlaces"`
	CompanyName      string `json:"companyName"`
	CompanyTaxID     string `json:"companyTaxId"`
	CompanyPhone     string `json:"companyPhone"`
	CompanyAddress   string `json:"companyAddress"`
}
