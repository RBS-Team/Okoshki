package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

type Repository interface {
	CreateAppointment(ctx context.Context, appt model.Appointment) error
	GetActiveAppointmentsByMaster(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]model.Appointment, error)
	GetAppointmentsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]model.Appointment, error)
	GetAppointmentsByMasterID(ctx context.Context, masterID uuid.UUID, start, end time.Time, status model.AppointmentStatus) ([]model.Appointment, error)
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status model.AppointmentStatus, masterNote *string) error
	DeleteManualBlock(ctx context.Context, id uuid.UUID, masterID uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &repository{db: db}
}
