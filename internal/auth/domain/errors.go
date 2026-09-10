package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("Invalid Email or Password")
	ErrInvalidToken       = errors.New("Invalid Token or session exired")
)
