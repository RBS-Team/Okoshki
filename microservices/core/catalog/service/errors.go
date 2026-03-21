package service

import (
	"errors"

	"github.com/RBS-Team/Okoshki/microservices/core/catalog/repository/postgres"
)

var (
	ErrNotFound        = errors.New("entity not found")
	ErrConflict        = errors.New("entity already exists")
	ErrInvalidTimezone = errors.New("invalid timezone provided")
)

func mapError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, postgres.ErrConflict):
		return ErrConflict
	default:
		return err
	}
}