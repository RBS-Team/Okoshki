package model

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID            uuid.UUID
	ClientID      uuid.UUID
	MasterID      uuid.UUID
	AppointmentID uuid.UUID
	Rating        int
	Comment       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
