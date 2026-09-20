package domain

import "errors"

var (
	ErrCustomerNotFound              = errors.New("customer not found")
	ErrCustomerDocumentExists        = errors.New("customer with this document number already exists")
	ErrCustomerHasAssociatedRecords  = errors.New("customer has associated records and cannot be deleted")
	ErrInvalidCustomerDocumentType   = errors.New("invalid document type: must be nit, ci, passport or other")
)

