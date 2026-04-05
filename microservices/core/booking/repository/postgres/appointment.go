package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) CreateAppointment(ctx context.Context, appt model.Appointment) error {
	const op = "booking.repository.postgres.CreateAppointment"

	query := `
		INSERT INTO appointments (
			id, client_id, master_id, service_id, start_at, end_at, 
			status, is_manual_block, client_comment, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		appt.ID, appt.ClientID, appt.MasterID, appt.ServiceID,
		appt.StartAt, appt.EndAt, appt.Status, appt.IsManualBlock,
		appt.ClientComment, appt.CreatedAt, appt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, mapAppointmentErrors(err))
	}

	return nil
}

func (r *Repository) GetActiveAppointmentsByMaster(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]model.Appointment, error) {
	const op = "booking.repository.postgres.GetActiveAppointmentsByMaster"

	query := `
		SELECT id, client_id, master_id, service_id, start_at, end_at, 
		       status, is_manual_block, client_comment, master_note, created_at, updated_at
		FROM appointments
		WHERE master_id = $1 
		  AND start_at >= $2 
		  AND start_at < $3
		  AND status IN ('pending', 'confirmed')
		ORDER BY start_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, masterID, start, end)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	var appointments []model.Appointment
	for rows.Next() {
		var a model.Appointment
		if err := rows.Scan(
			&a.ID, &a.ClientID, &a.MasterID, &a.ServiceID, &a.StartAt, &a.EndAt,
			&a.Status, &a.IsManualBlock, &a.ClientComment, &a.MasterNote,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		appointments = append(appointments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return appointments, nil
}
