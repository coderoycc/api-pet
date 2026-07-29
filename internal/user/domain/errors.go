package domain

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
	ErrInvalidInput  = errors.New("invalid input")
)
