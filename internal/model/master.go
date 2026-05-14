package model

import (
	"time"

	"github.com/google/uuid"
)

type Master struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CategoryID   uuid.UUID
	FirstName    string
	LastName     string
	Phone        string
	Address      string
	City         string
	Bio          *string
	AvatarURL    *string
	Timezone     string
	Lat          *float64
	Lon          *float64
	Rating       float64
	ReviewCount  int
	ReportsCount int
	IsBlocked    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
