package model

import (
	"time"

	"github.com/google/uuid"
)

// MasterSettings — пользовательские настройки мастера для расписания и записей.
// Связь 1‑к‑1 с masters; PK = MasterID.
type MasterSettings struct {
	MasterID        uuid.UUID
	SlotStepMinutes int
	LeadTimeMinutes int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
