package domain

import "errors"

var (
	ErrWarehouseNotFound = errors.New("warehouse not found")
	ErrWarehouseNameExists = errors.New("warehouse name already exists")
	ErrWarehouseInUse = errors.New("cannot delete warehouse because it is in use by purchases or inventory")
)
