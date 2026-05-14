package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkInterval — один рабочий промежуток мастера на конкретную дату.
// На одну дату может быть несколько интервалов (например, 10:00–14:00 и 16:00–20:00).
// StartTime/EndTime — локальные времена в таймзоне мастера, в формате "15:04".
type WorkInterval struct {
	ID        uuid.UUID
	MasterID  uuid.UUID
	WorkDate  time.Time
	StartTime string
	EndTime   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
