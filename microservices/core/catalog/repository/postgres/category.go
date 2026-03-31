package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	const op = "catalog.repository.postgres.GetCategoryByID"

	query := `
		SELECT id, parent_id, name, description, is_active, created_at, updated_at
		FROM category
		WHERE id = $1 AND is_active = true
	`

	var cat model.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cat.ID,
		&cat.ParentID,
		&cat.Name,
		&cat.Description,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	return &cat, nil
}

func (r *Repository) GetAllCategories(ctx context.Context) ([]model.Category, error) {
	const op = "catalog.repository.postgres.GetAllCategories"

	query := `
		SELECT id, parent_id, name, description, is_active, created_at, updated_at
		FROM category
		WHERE is_active = true
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	categories := make([]model.Category, 0)
	for rows.Next() {
		var cat model.Category
		if err := rows.Scan(
			&cat.ID,
			&cat.ParentID,
			&cat.Name,
			&cat.Description,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return categories, nil
}
