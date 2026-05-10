package model

import (
	"time"

	"github.com/google/uuid"
)

type PortfolioPhoto struct {
	ID         uuid.UUID
	MasterID   uuid.UUID
	ObjectName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
