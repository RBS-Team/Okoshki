package model

import (
	"time"

	"github.com/google/uuid"
)

type ServiceItem struct {
	ID              uuid.UUID
	MasterID        uuid.UUID
	CategoryID      uuid.UUID
	Title           string
	Address         string
	City            string
	Description     *string
	Price           int64
	DurationMinutes int
	IsActive        bool
	IsAutoConfirm   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
