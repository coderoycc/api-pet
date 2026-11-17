package domain

import "errors"

var (
	ErrCustomerNotFound             = errors.New("customer not found")
	ErrCustomerDocumentExists       = errors.New("customer with this document number already exists")
	ErrCustomerHasAssociatedRecords = errors.New("customer has associated records and cannot be deleted")

	ErrCustomerNameRequired           = errors.New("name is required")
	ErrCustomerDocumentTypeRequired   = errors.New("documentType is required")
	ErrCustomerDocumentNumberRequired = errors.New("documentNumber is required")
	ErrInvalidCustomerDocumentType    = errors.New("invalid document type: must be nit, ci, passport or other")
	ErrInvalidCustomerStatus          = errors.New("invalid status: must be active, inactive or suspended")
	ErrInvalidCustomerEmail           = errors.New("invalid email format")
)
