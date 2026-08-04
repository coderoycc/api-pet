package domain

import "errors"

var (
	ErrCustomerNotFound       = errors.New("customer not found")
	ErrCustomerDocumentExists = errors.New("customer with this document number already exists")
)
