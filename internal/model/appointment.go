package model

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"   // Ожидает подтверждения мастером
	StatusConfirmed AppointmentStatus = "confirmed" // Подтверждена (или авто-подтверждена)
	StatusRejected  AppointmentStatus = "rejected"  // Отклонена мастером
	StatusCancelled AppointmentStatus = "cancelled" // Отменена клиентом
	StatusCompleted AppointmentStatus = "completed" // Услуга оказана (можно оставлять отзыв)
)

type Appointment struct {
	ID            uuid.UUID
	ClientID      uuid.UUID
	MasterID      uuid.UUID
	ServiceID     uuid.UUID
	StartAt       time.Time
	EndAt         time.Time
	Status        AppointmentStatus
	IsManualBlock bool
	ClientComment *string
	MasterNote    *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
