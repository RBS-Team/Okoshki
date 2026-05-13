package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *Repository) CreateMaster(ctx context.Context, master model.Master) error {
	const op = "catalog.repository.postgres.CreateMaster"

	query := `
		INSERT INTO masters (
			id, user_id, category_id, first_name, last_name, phone, address, city, bio, avatar_url, timezone,
			lat, lon, rating, review_count, reports_count, is_blocked,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.db.ExecContext(ctx, query,
		master.ID,
		master.UserID,
		master.CategoryID,
		master.FirstName,
		master.LastName,
		master.Phone,
		master.Address,
		master.City,
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
	const op = "users.repository.postgres.GetMasterByID"

	query := `
		SELECT id, user_id, category_id, first_name, last_name, phone, address, city, bio, avatar_url, timezone,
			lat, lon, rating, review_count, reports_count, is_blocked,
			created_at, updated_at
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
	const op = "users.repository.postgres.GetAllMasters"

	query := `
		SELECT id, user_id, category_id, first_name, last_name, phone, address, city, bio, avatar_url, timezone,
			lat, lon, rating, review_count, reports_count, is_blocked,
			created_at, updated_at
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
			&m.ID, &m.UserID, &m.CategoryID, &m.FirstName, &m.LastName, &m.Phone,
			&m.Address, &m.City, &m.Bio, &m.AvatarURL, &m.Timezone, &m.Lat,
			&m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount, &m.IsBlocked,
			&m.CreatedAt, &m.UpdatedAt,
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
	const op = "users.repository.postgres.GetMastersByIDs"

	if len(ids) == 0 {
		return []model.Master{}, nil
	}

	query := `
		SELECT id, user_id, category_id, first_name, last_name, phone, address, city, bio, avatar_url, timezone,
			lat, lon, rating, review_count, reports_count, is_blocked,
			created_at, updated_at
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
			&m.ID, &m.UserID, &m.CategoryID, &m.FirstName, &m.LastName, &m.Phone,
			&m.Address, &m.City, &m.Bio, &m.AvatarURL, &m.Timezone, &m.Lat,
			&m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount, &m.IsBlocked,
			&m.CreatedAt, &m.UpdatedAt,
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

func (r *Repository) GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*model.Master, error) {
	const op = "users.repository.postgres.GetMasterByUserID"

	query := `
		SELECT id, user_id, category_id, first_name, last_name, phone, address, city, bio, avatar_url, timezone,
			lat, lon, rating, review_count, reports_count, is_blocked,
			created_at, updated_at
		FROM masters
		WHERE user_id = $1 AND is_blocked = false
	`

	master, err := r.selectMaster(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return master, nil
}

func (r *Repository) selectMaster(ctx context.Context, query string, args ...any) (*model.Master, error) {
	var m model.Master

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&m.ID, &m.UserID, &m.CategoryID, &m.FirstName, &m.LastName, &m.Phone,
		&m.Address, &m.City, &m.Bio, &m.AvatarURL, &m.Timezone,
		&m.Lat, &m.Lon, &m.Rating, &m.ReviewCount, &m.ReportsCount,
		&m.IsBlocked, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &m, nil
}

func (r *Repository) GetMastersByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.Master, error) {
	const op = "users.repository.postgres.GetMastersByCategoryID"

	query := `
		SELECT id, user_id, category_id, first_name, last_name, phone, address, city, bio, avatar_url, timezone,
			lat, lon, rating, review_count, reports_count, is_blocked,
			created_at, updated_at
		FROM masters
		WHERE is_blocked = false
		  AND category_id = $1
		ORDER BY rating DESC, created_at DESC
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
			&m.ID, &m.UserID, &m.CategoryID, &m.FirstName, &m.LastName, &m.Phone,
			&m.Address, &m.City, &m.Bio, &m.AvatarURL, &m.Timezone,
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

func (r *Repository) UpdateMasterAvatarURL(ctx context.Context, id uuid.UUID, objectName string) error {
	const op = "users.repository.postgres.UpdateMasterAvatarURL"

	_, err := r.db.ExecContext(ctx,
		`UPDATE masters SET avatar_url = $1 WHERE id = $2`,
		objectName, id,
	)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func handleMasterPostgresError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return domain.ErrConflict
		}
	}

	return err
}
