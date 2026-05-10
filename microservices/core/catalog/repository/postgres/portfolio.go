package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) SavePortfolioPhotos(ctx context.Context, photos []model.PortfolioPhoto) error {
	const op = "catalog.repository.postgres.SavePortfolioPhotos"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[%s]: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO master_portfolio_photos (id, master_id, object_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("[%s]: prepare: %w", op, err)
	}
	defer stmt.Close()

	for _, p := range photos {
		if _, err := stmt.ExecContext(ctx, p.ID, p.MasterID, p.ObjectName, p.CreatedAt, p.UpdatedAt); err != nil {
			return fmt.Errorf("[%s]: exec: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("[%s]: commit: %w", op, err)
	}

	return nil
}

func (r *Repository) GetPortfolioPhotosByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.PortfolioPhoto, error) {
	const op = "catalog.repository.postgres.GetPortfolioPhotosByMasterID"

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, master_id, object_name, created_at, updated_at
		FROM master_portfolio_photos
		WHERE master_id = $1
		ORDER BY created_at DESC
	`, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query: %w", op, err)
	}
	defer rows.Close()

	photos := make([]model.PortfolioPhoto, 0)
	for rows.Next() {
		var p model.PortfolioPhoto
		if err := rows.Scan(&p.ID, &p.MasterID, &p.ObjectName, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("[%s]: scan: %w", op, err)
		}
		photos = append(photos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows: %w", op, err)
	}

	return photos, nil
}
