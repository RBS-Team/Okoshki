package model

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FirstName string
	LastName  string
	Phone     string
	AvatarURL *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
