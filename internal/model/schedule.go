package model

import (
	"time"

	"github.com/google/uuid"
)

type WorkingHours struct {
	ID        uuid.UUID
	MasterID  uuid.UUID
	DayOfWeek int
	StartTime *string
	EndTime   *string
	IsDayOff  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ScheduleException struct {
	ID            uuid.UUID
	MasterID      uuid.UUID
	ExceptionDate time.Time
	StartTime     *string
	EndTime       *string
	IsWorking     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
