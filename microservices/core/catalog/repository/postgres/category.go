package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *repository) UpdateCategoryAvatarURL(ctx context.Context, id uuid.UUID, objectName string) error {
	const op = "catalog.repository.postgres.UpdateCategoryAvatarURL"

	_, err := r.db.ExecContext(ctx,
		`UPDATE category SET avatar_url = $1 WHERE id = $2`,
		objectName, id,
	)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func (r *repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	const op = "catalog.repository.postgres.GetCategoryByID"

	query := `
		SELECT c.id, c.name, c.description, c.avatar_url, c.is_active, c.created_at, c.updated_at,
		       COUNT(m.id) AS masters_count
		FROM category c
		LEFT JOIN masters m ON m.category_id = c.id AND m.is_blocked = false
		WHERE c.id = $1 AND c.is_active = true
		GROUP BY c.id
	`

	var cat model.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Description,
		&cat.AvatarURL,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
		&cat.MastersCount,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapErrors(err))
	}

	return &cat, nil
}

func (r *repository) GetAllCategories(ctx context.Context) ([]model.Category, error) {
	const op = "catalog.repository.postgres.GetAllCategories"

	query := `
		SELECT c.id, c.name, c.description, c.avatar_url, c.is_active, c.created_at, c.updated_at,
		       COUNT(m.id) AS masters_count
		FROM category c
		LEFT JOIN masters m ON m.category_id = c.id AND m.is_blocked = false
		WHERE c.is_active = true
		GROUP BY c.id
		ORDER BY c.name ASC
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
			&cat.Name,
			&cat.Description,
			&cat.AvatarURL,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
			&cat.MastersCount,
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
