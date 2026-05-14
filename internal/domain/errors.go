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

	// Расписание и слоты.
	ErrSlotNotAvailable      = errors.New("requested slot is not available")
	ErrLeadTimeViolation     = errors.New("lead time violation: too close to start")
	ErrIntervalOverlap       = errors.New("work interval overlaps with existing one")
	ErrIntervalHasAppointments = errors.New("work interval contains existing appointments")
)
