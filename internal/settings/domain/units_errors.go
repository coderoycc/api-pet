package domain

import "errors"

var (
	ErrUnitNotFound       = errors.New("unit of measure not found")
	ErrUnitAbbrevExists   = errors.New("unit of measure abbreviation already exists")
)
