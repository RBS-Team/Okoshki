package model

import (
	"time"

	"github.com/google/uuid"
)

type ServiceItem struct {
	ID                  uuid.UUID
	MasterID            uuid.UUID
	CategoryID          uuid.UUID
	Title               string
	Address				string
	Description         *string
	Price               float64
	DurationMinutes     int
	BufferBeforeMinutes int
	BufferAfterMinutes  int
	IsActive            bool
	IsAutoConfirm       bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
