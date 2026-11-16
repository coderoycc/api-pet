package application

import (
	"net/mail"
	"strings"

	"api-go/internal/customers/domain"
)

var validDocumentTypes = map[string]bool{
	domain.DocTypeNIT:      true,
	domain.DocTypeCI:       true,
	domain.DocTypePassport: true,
	domain.DocTypeOther:    true,
}

var validStatuses = map[string]bool{
	domain.StatusActive:    true,
	domain.StatusInactive:  true,
	domain.StatusSuspended: true,
}

// customerInput representa los campos comunes validables entre creación y actualización.
type customerInput struct {
	Name           string
	DocumentType   string
	DocumentNumber string
	Email          *string
	Phone          *string
	MobilePhone    *string
	Address        *string
	City           *string
	Department     *string
	BusinessName   *string
	TaxCategory    *string
	Notes          *string
	Status         string
}

func trimOptionalString(val *string) *string {
	if val == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// normalizeAndValidateCustomer sanitiza los campos y valida las reglas de negocio.
func normalizeAndValidateCustomer(input *customerInput, defaultStatus string) error {
	input.Name = strings.TrimSpace(input.Name)
	input.DocumentType = strings.ToLower(strings.TrimSpace(input.DocumentType))
	input.DocumentNumber = strings.TrimSpace(input.DocumentNumber)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))

	input.Email = trimOptionalString(input.Email)
	input.Phone = trimOptionalString(input.Phone)
	input.MobilePhone = trimOptionalString(input.MobilePhone)
	input.Address = trimOptionalString(input.Address)
	input.City = trimOptionalString(input.City)
	input.Department = trimOptionalString(input.Department)
	input.BusinessName = trimOptionalString(input.BusinessName)
	input.TaxCategory = trimOptionalString(input.TaxCategory)
	input.Notes = trimOptionalString(input.Notes)

	if input.Name == "" {
		return domain.ErrCustomerNameRequired
	}
	if input.DocumentType == "" {
		return domain.ErrCustomerDocumentTypeRequired
	}
	if !validDocumentTypes[input.DocumentType] {
		return domain.ErrInvalidCustomerDocumentType
	}
	if input.DocumentNumber == "" {
		return domain.ErrCustomerDocumentNumberRequired
	}
	if input.Email != nil {
		if _, err := mail.ParseAddress(*input.Email); err != nil {
			return domain.ErrInvalidCustomerEmail
		}
	}

	if input.Status == "" {
		input.Status = defaultStatus
	}
	if input.Status != "" && !validStatuses[input.Status] {
		return domain.ErrInvalidCustomerStatus
	}

	return nil
}
