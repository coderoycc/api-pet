package domain

import "errors"

var (
	ErrPurchaseNotFound       = errors.New("purchase not found")
	ErrCannotReceiveCancelled = errors.New("cannot receive a cancelled purchase")
	ErrAlreadyReceived        = errors.New("purchase is already received")
	ErrCannotCancelReceived   = errors.New("cannot cancel an already received purchase")
)
