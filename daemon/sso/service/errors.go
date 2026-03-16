package service

import (
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrAppNotFound        = errors.New("application not found")
	ErrInternal           = errors.New("internal error")
)
