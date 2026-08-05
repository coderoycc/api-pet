package domain

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrDuplicateSKU    = errors.New("product with this SKU already exists")
	ErrBatchNotFound   = errors.New("product batch not found")
	ErrInvalidQuantity = errors.New("invalid product quantity")
)
