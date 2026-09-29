package domain

import "errors"

var (
	ErrSupplierNotFound    = errors.New("supplier not found")
	ErrSupplierTaxIdExists = errors.New("supplier tax id already exists")
	ErrSupplierInUse       = errors.New("cannot delete supplier because it is in use by products, purchases or inventory")
)
