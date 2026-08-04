package domain

import "errors"

var (
	ErrWarehouseNotFound = errors.New("warehouse not found")
	ErrWarehouseNameExists = errors.New("warehouse name already exists")
)
