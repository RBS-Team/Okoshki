package postgres

import "errors"

var (
	ErrNotFound = errors.New("user not found")
	ErrConflict = errors.New("user already exists")
	ErrInternal = errors.New("internal")
)