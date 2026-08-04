package application

type CustomerCreateDTO struct {
	Name           string  `json:"name" validate:"required"`
	DocumentType   string  `json:"documentType" validate:"required"`
	DocumentNumber string  `json:"documentNumber" validate:"required"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	MobilePhone    *string `json:"mobilePhone"`
	Address        *string `json:"address"`
	City           *string `json:"city"`
	Department     *string `json:"department"`
	BusinessName   *string `json:"businessName"`
	TaxCategory    *string `json:"taxCategory"`
	Notes          *string `json:"notes"`
	Status         string  `json:"status"` // active, inactive, suspended
}

type CustomerUpdateDTO struct {
	Name           string  `json:"name" validate:"required"`
	DocumentType   string  `json:"documentType" validate:"required"`
	DocumentNumber string  `json:"documentNumber" validate:"required"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	MobilePhone    *string `json:"mobilePhone"`
	Address        *string `json:"address"`
	City           *string `json:"city"`
	Department     *string `json:"department"`
	BusinessName   *string `json:"businessName"`
	TaxCategory    *string `json:"taxCategory"`
	Notes          *string `json:"notes"`
	Status         string  `json:"status"`
}
