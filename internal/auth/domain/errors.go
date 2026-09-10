package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid token or session expired")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrUserDeleted        = errors.New("user account is deleted")
)
