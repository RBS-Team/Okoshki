package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
)

// GetMasterSettings возвращает настройки мастера. Если строки нет — domain.ErrNotFound.
// Решение «что подставить вместо» принимается на уровне сервиса.
func (r *repository) GetMasterSettings(ctx context.Context, masterID uuid.UUID) (*model.MasterSettings, error) {
	const op = "catalog.repository.postgres.GetMasterSettings"

	query := `
		SELECT master_id, slot_step_minutes, lead_time_minutes, created_at, updated_at
		FROM master_settings
		WHERE master_id = $1
	`

	var s model.MasterSettings
	err := r.db.QueryRowContext(ctx, query, masterID).Scan(
		&s.MasterID, &s.SlotStepMinutes, &s.LeadTimeMinutes,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("[%s]: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
	}

	return &s, nil
}

// UpsertMasterSettings создаёт строку настроек или обновляет её.
// Поля с nil НЕ перезаписываются (для частичного update).
// При вставке используются дефолты столбцов БД (step=30, lead=0).
func (r *repository) UpsertMasterSettings(ctx context.Context, masterID uuid.UUID, slotStep, leadTime *int) error {
	const op = "catalog.repository.postgres.UpsertMasterSettings"

	query := `
		INSERT INTO master_settings (master_id, slot_step_minutes, lead_time_minutes)
		VALUES ($1, COALESCE($2, 30), COALESCE($3, 0))
		ON CONFLICT (master_id) DO UPDATE SET
			slot_step_minutes = COALESCE($2, master_settings.slot_step_minutes),
			lead_time_minutes = COALESCE($3, master_settings.lead_time_minutes),
			updated_at        = CURRENT_TIMESTAMP
	`

	if _, err := r.db.ExecContext(ctx, query, masterID, slotStep, leadTime); err != nil {
		return fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	return nil
}
