package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) CreateServiceItem(ctx context.Context, item model.ServiceItem) error {
	const op = "catalog.repository.postgres.CreateServiceItem"

	query := `
		INSERT INTO master_services (
			id, master_id, category_id, title, description, price, 
			duration_minutes, buffer_before_minutes, buffer_after_minutes, 
			is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.MasterID,
		item.CategoryID,
		item.Title,
		item.Description,
		item.Price,
		item.DurationMinutes,
		item.BufferBeforeMinutes,
		item.BufferAfterMinutes,
		item.IsActive,
		item.CreatedAt,
		item.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func (r *Repository) GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.ServiceItem, error) {
	const op = "catalog.repository.postgres.GetServiceItemsByMasterID"

	query := `
		SELECT id, master_id, category_id, title, description, price, 
		       duration_minutes, buffer_before_minutes, buffer_after_minutes, 
		       is_active, created_at, updated_at
		FROM master_services
		WHERE master_id = $1 AND is_active = true
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	items := make([]model.ServiceItem, 0)
	for rows.Next() {
		var item model.ServiceItem
		if err := rows.Scan(
			&item.ID, &item.MasterID, &item.CategoryID, &item.Title, &item.Description,
			&item.Price, &item.DurationMinutes, &item.BufferBeforeMinutes,
			&item.BufferAfterMinutes, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return items, nil
}