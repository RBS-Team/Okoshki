package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) UpsertWorkingHours(ctx context.Context, masterID uuid.UUID, hours []model.WorkingHours) error {
	const op = "catalog.repository.postgres.UpsertWorkingHours"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[%s]: failed to begin tx: %w", op, err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO master_working_hours (master_id, day_of_week, start_time, end_time, is_day_off)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (master_id, day_of_week) 
		DO UPDATE SET 
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			is_day_off = EXCLUDED.is_day_off,
			updated_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("[%s]: failed to prepare stmt: %w", op, err)
	}
	defer stmt.Close()

	for _, h := range hours {
		_, err := stmt.ExecContext(ctx, masterID, h.DayOfWeek, h.StartTime, h.EndTime, h.IsDayOff)
		if err != nil {
			return fmt.Errorf("[%s]: failed to execute stmt for day %d: %w", op, h.DayOfWeek, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("[%s]: failed to commit tx: %w", op, err)
	}

	return nil
}

func (r *Repository) GetWorkingHoursByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.WorkingHours, error) {
	const op = "catalog.repository.postgres.GetWorkingHoursByMasterID"

	query := `
		SELECT id, master_id, day_of_week, start_time::text, end_time::text, is_day_off, created_at, updated_at
		FROM master_working_hours
		WHERE master_id = $1
		ORDER BY day_of_week ASC
	`

	rows, err := r.db.QueryContext(ctx, query, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	hours := make([]model.WorkingHours, 0)
	for rows.Next() {
		var h model.WorkingHours
		if err := rows.Scan(
			&h.ID, &h.MasterID, &h.DayOfWeek, &h.StartTime, &h.EndTime,
			&h.IsDayOff, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		hours = append(hours, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return hours, nil
}

func (r *Repository) CreateScheduleException(ctx context.Context, exc model.ScheduleException) error {
	const op = "catalog.repository.postgres.CreateScheduleException"

	query := `
		INSERT INTO master_schedule_exceptions (
			id, master_id, exception_date, start_time, end_time, is_working, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		exc.ID, exc.MasterID, exc.ExceptionDate, exc.StartTime, exc.EndTime,
		exc.IsWorking, exc.CreatedAt, exc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	return nil
}

func (r *Repository) GetScheduleExceptionByID(ctx context.Context, masterID, exceptionID uuid.UUID) (*model.ScheduleException, error) {
	const op = "catalog.repository.postgres.GetScheduleExceptionByID"

	query := `
		SELECT id, master_id, exception_date, start_time::text, end_time::text, is_working, created_at, updated_at
		FROM master_schedule_exceptions
		WHERE id = $1 AND master_id = $2
	`

	var exc model.ScheduleException
	err := r.db.QueryRowContext(ctx, query, exceptionID, masterID).Scan(
		&exc.ID, &exc.MasterID, &exc.ExceptionDate, &exc.StartTime, &exc.EndTime,
		&exc.IsWorking, &exc.CreatedAt, &exc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
	}

	return &exc, nil
}

func (r *Repository) GetScheduleExceptions(ctx context.Context, masterID uuid.UUID, startDate, endDate time.Time) ([]model.ScheduleException, error) {
	const op = "catalog.repository.postgres.GetScheduleExceptions"

	query := `
		SELECT id, master_id, exception_date, start_time::text, end_time::text, is_working, created_at, updated_at
		FROM master_schedule_exceptions
		WHERE master_id = $1 AND exception_date >= $2 AND exception_date <= $3
		ORDER BY exception_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, masterID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	exceptions := make([]model.ScheduleException, 0)
	for rows.Next() {
		var exc model.ScheduleException
		if err := rows.Scan(
			&exc.ID, &exc.MasterID, &exc.ExceptionDate, &exc.StartTime, &exc.EndTime,
			&exc.IsWorking, &exc.CreatedAt, &exc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		exceptions = append(exceptions, exc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return exceptions, nil
}

type UpdateScheduleExceptionInput struct {
	StartTime *string
	EndTime   *string
	IsWorking *bool
}

func (r *Repository) UpdateScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID, upd UpdateScheduleExceptionInput) error {
	const op = "catalog.repository.postgres.UpdateScheduleException"

	setParts := []string{}
	args := []interface{}{}
	argID := 1

	if upd.StartTime != nil {
		setParts = append(setParts, fmt.Sprintf("start_time = $%d", argID))
		args = append(args, *upd.StartTime)
		argID++
	}
	if upd.EndTime != nil {
		setParts = append(setParts, fmt.Sprintf("end_time = $%d", argID))
		args = append(args, *upd.EndTime)
		argID++
	}
	if upd.IsWorking != nil {
		setParts = append(setParts, fmt.Sprintf("is_working = $%d", argID))
		args = append(args, *upd.IsWorking)
		argID++
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, exceptionID, masterID)

	query := fmt.Sprintf(
		`UPDATE master_schedule_exceptions SET %s, updated_at = CURRENT_TIMESTAMP WHERE id = $%d AND master_id = $%d`,
		strings.Join(setParts, ", "), argID, argID+1,
	)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("[%s]: could not get rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID) error {
	const op = "catalog.repository.postgres.DeleteScheduleException"

	query := `DELETE FROM master_schedule_exceptions WHERE id = $1 AND master_id = $2`

	res, err := r.db.ExecContext(ctx, query, exceptionID, masterID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("[%s]: could not get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
