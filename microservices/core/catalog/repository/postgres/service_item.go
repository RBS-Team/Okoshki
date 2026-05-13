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
			id, master_id, category_id, title, address, city, description, price,
			duration_minutes, is_active, is_auto_confirm, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.MasterID,
		item.CategoryID,
		item.Title,
		item.Address,
		item.City,
		item.Description,
		item.Price,
		item.DurationMinutes,
		item.IsActive,
		item.IsAutoConfirm,
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
		SELECT id, master_id, category_id, title, address, city, description, price,
		       duration_minutes, is_active, is_auto_confirm, created_at, updated_at
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
			&item.ID, &item.MasterID, &item.CategoryID, &item.Title, &item.Address, &item.City, &item.Description,
			&item.Price, &item.DurationMinutes, &item.IsActive, &item.IsAutoConfirm,
			&item.CreatedAt, &item.UpdatedAt,
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

func (r *Repository) GetServicesByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.ServiceItem, error) {
	const op = "catalog.repository.postgres.GetServicesByCategoryID"

	query := `
		SELECT id, master_id, category_id, title, address, city, description, price,
		       duration_minutes, is_active, is_auto_confirm, created_at, updated_at
		FROM master_services
		WHERE is_active = true AND category_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	items := make([]model.ServiceItem, 0)
	for rows.Next() {
		var s model.ServiceItem
		if err := rows.Scan(
			&s.ID, &s.MasterID, &s.CategoryID, &s.Title, &s.Address, &s.City, &s.Description,
			&s.Price, &s.DurationMinutes, &s.IsActive, &s.IsAutoConfirm,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		items = append(items, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return items, nil
}

func (r *Repository) GetServiceItemByID(ctx context.Context, id uuid.UUID) (*model.ServiceItem, error) {
	const op = "catalog.repository.postgres.GetServiceItemByID"

	query := `
		SELECT id, master_id, category_id, title, address, city, description, price,
		       duration_minutes, is_active, is_auto_confirm, created_at, updated_at
		FROM master_services
		WHERE id = $1 AND is_active = true
	`

	var item model.ServiceItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.MasterID, &item.CategoryID, &item.Title, &item.Address, &item.City, &item.Description,
		&item.Price, &item.DurationMinutes, &item.IsActive, &item.IsAutoConfirm,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	return &item, nil
}
