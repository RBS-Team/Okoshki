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

func (r *Repository) CreateUser(ctx context.Context, user model.User) error {
	const op = "auth.repository.postgres.CreateUser"

	query := `
		INSERT INTO "user" (user_id, email, password_hash, role, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, handlePostgresError(err))
	}

	return nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	const op = "auth.repository.postgres.GetUserByEmail"

	query := `
		SELECT user_id, email, password_hash, role, avatar_url, created_at, updated_at
		FROM "user" 
		WHERE email = $1
	`

	user, err := r.selectUser(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const op = "auth.repository.postgres.GetUserByID"

	query := `
		SELECT user_id, email, password_hash, role, avatar_url, created_at, updated_at
		FROM "user" 
		WHERE user_id = $1
	`

	user, err := r.selectUser(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *Repository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error) {
	const op = "auth.repository.postgres.GetUsersByIDs"

	if len(ids) == 0 {
		return []model.User{}, nil
	}

	query := `
		SELECT user_id, email, password_hash, role, avatar_url, created_at, updated_at
		FROM "user" 
		WHERE user_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return users, nil
}

func (r *Repository) selectUser(ctx context.Context, query string, args ...interface{}) (*model.User, error) {
	var user model.User

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func handlePostgresError(err error) error {
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
