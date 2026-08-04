package domain

import "errors"

var (
	ErrSupplierNotFound    = errors.New("supplier not found")
	ErrSupplierTaxIdExists = errors.New("supplier tax id already exists")
)
