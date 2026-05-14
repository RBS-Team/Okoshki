package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	AvatarURL    *string
	MastersCount int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
