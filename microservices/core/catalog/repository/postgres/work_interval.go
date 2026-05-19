package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
)

// CreateWorkInterval вставляет один интервал.
// При пересечении с существующим возвращает domain.ErrIntervalOverlap (через mapErrors).
func (r *repository) CreateWorkInterval(ctx context.Context, wi model.WorkInterval) error {
	const op = "catalog.repository.postgres.CreateWorkInterval"

	query := `
		INSERT INTO master_work_intervals (id, master_id, work_date, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
	`

	if _, err := r.db.ExecContext(ctx, query, wi.ID, wi.MasterID, wi.WorkDate, wi.StartTime, wi.EndTime); err != nil {
		return fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	return nil
}

// DeleteWorkInterval удаляет интервал по id, проверяя принадлежность мастеру.
// Если интервал не найден — domain.ErrNotFound.
func (r *repository) DeleteWorkInterval(ctx context.Context, masterID, intervalID uuid.UUID) error {
	const op = "catalog.repository.postgres.DeleteWorkInterval"

	res, err := r.db.ExecContext(ctx, `DELETE FROM master_work_intervals WHERE id = $1 AND master_id = $2`, intervalID, masterID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("[%s]: rows affected: %w", op, err)
	}
	if rows == 0 {
		return fmt.Errorf("[%s]: %w", op, domain.ErrNotFound)
	}

	return nil
}

// GetWorkIntervalByID возвращает интервал по id с проверкой мастера.
func (r *repository) GetWorkIntervalByID(ctx context.Context, masterID, intervalID uuid.UUID) (*model.WorkInterval, error) {
	const op = "catalog.repository.postgres.GetWorkIntervalByID"

	query := `
		SELECT id, master_id, work_date, start_time::text, end_time::text, created_at, updated_at
		FROM master_work_intervals
		WHERE id = $1 AND master_id = $2
	`

	var wi model.WorkInterval
	err := r.db.QueryRowContext(ctx, query, intervalID, masterID).Scan(&wi.ID, &wi.MasterID, &wi.WorkDate, &wi.StartTime, &wi.EndTime, &wi.CreatedAt, &wi.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("[%s]: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
	}

	return &wi, nil
}

// GetWorkIntervalsByMasterRange возвращает все интервалы мастера в [from, to] включительно.
// Сортировка: дата ASC, затем start_time ASC.
func (r *repository) GetWorkIntervalsByMasterRange(ctx context.Context, masterID uuid.UUID, from, to time.Time) ([]model.WorkInterval, error) {
	const op = "catalog.repository.postgres.GetWorkIntervalsByMasterRange"

	query := `
		SELECT id, master_id, work_date, start_time::text, end_time::text, created_at, updated_at
		FROM master_work_intervals
		WHERE master_id = $1 AND work_date >= $2 AND work_date <= $3
		ORDER BY work_date ASC, start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, masterID, from, to)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	intervals := make([]model.WorkInterval, 0)
	for rows.Next() {
		var wi model.WorkInterval
		if err := rows.Scan(&wi.ID, &wi.MasterID, &wi.WorkDate, &wi.StartTime, &wi.EndTime, &wi.CreatedAt, &wi.UpdatedAt); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		intervals = append(intervals, wi)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return intervals, nil
}

// ReplaceWorkIntervalsForDate атомарно заменяет все интервалы мастера на дату workDate.
// Пустой intervals = удалить все существующие на эту дату (мастер не работает).
func (r *repository) ReplaceWorkIntervalsForDate(ctx context.Context, masterID uuid.UUID, workDate time.Time, intervals []model.WorkInterval) error {
	const op = "catalog.repository.postgres.ReplaceWorkIntervalsForDate"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[%s]: begin tx: %w", op, err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM master_work_intervals WHERE master_id = $1 AND work_date = $2`, masterID, workDate); err != nil {
		return fmt.Errorf("[%s]: delete: %w", op, mapErrors(err))
	}

	if len(intervals) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO master_work_intervals (id, master_id, work_date, start_time, end_time)
			VALUES ($1, $2, $3, $4, $5)
		`)
		if err != nil {
			return fmt.Errorf("[%s]: prepare: %w", op, err)
		}
		defer stmt.Close()

		for _, wi := range intervals {
			if _, err := stmt.ExecContext(ctx, wi.ID, wi.MasterID, wi.WorkDate, wi.StartTime, wi.EndTime); err != nil {
				return fmt.Errorf("[%s]: insert: %w", op, mapErrors(err))
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("[%s]: commit: %w", op, err)
	}

	return nil
}

// HasActiveAppointmentsInRange проверяет, есть ли активные (pending/confirmed) записи мастера,
// которые целиком или частично попадают в [startUTC, endUTC).
// Используется перед удалением/заменой интервала, чтобы не оставить «висящие» записи.
func (r *repository) HasActiveAppointmentsInRange(ctx context.Context, masterID uuid.UUID, startUTC, endUTC time.Time) (bool, error) {
	const op = "catalog.repository.postgres.HasActiveAppointmentsInRange"

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM appointments
			WHERE master_id = $1
			  AND status IN ('pending', 'confirmed')
			  AND start_at < $3
			  AND end_at   > $2
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, masterID, startUTC, endUTC).Scan(&exists); err != nil {
		return false, fmt.Errorf("[%s]: scan failed: %w", op, err)
	}

	return exists, nil
}
