package domain

import (
	"errors"
)

var (
	ErrNotFound        = errors.New("entity not found")
	ErrConflict        = errors.New("entity already exists")
	ErrForbidden       = errors.New("access denied")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidTimezone = errors.New("invalid timezone provided")
	ErrInternal        = errors.New("internal")
)
