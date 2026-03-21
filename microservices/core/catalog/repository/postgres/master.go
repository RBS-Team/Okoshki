package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) CreateMaster(ctx context.Context, master model.Master) error {
	const op = "catalog.repository.postgres.CreateMaster"

	query := `
		INSERT INTO masters (
			id, user_id, name, bio, avatar_url, timezone, 
			lat, lon, rating, review_count, reports_count, is_blocked, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		master.ID,
		master.UserID,
		master.Name,
		master.Bio,
		master.AvatarURL,
		master.Timezone,
		master.Lat,
		master.Lon,
		master.Rating,
		master.ReviewCount,
		master.ReportsCount,
		master.IsBlocked,
		master.CreatedAt,
		master.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("[%s]: %w", op, handleMasterPostgresError(err))
	}

	return nil
}

func (r *Repository) GetMasterByID(ctx context.Context, id uuid.UUID) (*model.Master, error) {
	const op = "catalog.repository.postgres.GetMasterByID"

	query := `
		SELECT id, user_id, name, bio, avatar_url, timezone, lat, lon, 
		       rating, review_count, reports_count, is_blocked, created_at, updated_at
		FROM masters
		WHERE id = $1 AND is_blocked = false
	`

	master, err := r.selectMaster(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return master, nil
}

func (r *Repository) GetAllMasters(ctx context.Context, limit, offset uint64) ([]model.Master, error) {
	const op = "catalog.repository.postgres.GetAllMasters"

	query := `
		SELECT id, user_id, name, bio, avatar_url, timezone, lat, lon, 
		       rating, review_count, reports_count, is_blocked, created_at, updated_at
		FROM masters
		WHERE is_blocked = false
		ORDER BY rating DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	masters := make([]model.Master, 0)
	for rows.Next() {
		var m model.Master
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Bio, &m.AvatarURL, &m.Timezone,
			&m.Lat, &m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount,
			&m.IsBlocked, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		masters = append(masters, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return masters, nil
}

// GetMastersByCategoryID ищет мастеров, оказывающих услуги в указанной категории или любых её подкатегориях.
func (r *Repository) GetMastersByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.Master, error) {
	const op = "catalog.repository.postgres.GetMastersByCategoryID"

	query := `
		WITH RECURSIVE cat_tree AS (
			SELECT id FROM category WHERE id = $1 AND is_active = true
			UNION ALL
			SELECT c.id FROM category c
			INNER JOIN cat_tree ct ON c.parent_id = ct.id
			WHERE c.is_active = true
		)
		SELECT m.id, m.user_id, m.name, m.bio, m.avatar_url, m.timezone, m.lat, m.lon, 
		       m.rating, m.review_count, m.reports_count, m.is_blocked, m.created_at, m.updated_at
		FROM masters m
		WHERE m.is_blocked = false
		  AND EXISTS (
			  SELECT 1 FROM master_services ms
			  WHERE ms.master_id = m.id
			    AND ms.is_active = true
			    AND ms.category_id IN (SELECT id FROM cat_tree)
		  )
		ORDER BY m.rating DESC, m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	masters := make([]model.Master, 0)
	for rows.Next() {
		var m model.Master
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Bio, &m.AvatarURL, &m.Timezone,
			&m.Lat, &m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount,
			&m.IsBlocked, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		masters = append(masters, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return masters, nil
}

func (r *Repository) GetMastersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Master, error) {
	const op = "catalog.repository.postgres.GetMastersByIDs"

	if len(ids) == 0 {
		return []model.Master{}, nil
	}

	query := `
		SELECT id, user_id, name, bio, avatar_url, timezone, lat, lon, 
		       rating, review_count, reports_count, is_blocked, created_at, updated_at
		FROM masters
		WHERE id = ANY($1) AND is_blocked = false
	`

	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	masters := make([]model.Master, 0, len(ids))
	for rows.Next() {
		var m model.Master
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Bio, &m.AvatarURL, &m.Timezone,
			&m.Lat, &m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount,
			&m.IsBlocked, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		masters = append(masters, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return masters, nil
}

func (r *Repository) selectMaster(ctx context.Context, query string, args ...interface{}) (*model.Master, error) {
	var m model.Master

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Bio, &m.AvatarURL, &m.Timezone,
		&m.Lat, &m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount,
		&m.IsBlocked, &m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &m, nil
}

func handleMasterPostgresError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrConflict
		}
	}

	return err
}
