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

func (r *repository) CreateClient(ctx context.Context, client model.Client) error {
	const op = "users.repository.postgres.CreateClient"

	query := `
		INSERT INTO clients (id, user_id, first_name, last_name, phone, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		client.ID,
		client.UserID,
		client.FirstName,
		client.LastName,
		client.Phone,
		client.AvatarURL,
		client.CreatedAt,
		client.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, handleMasterPostgresError(err))
	}

	return nil
}

func (r *repository) GetClientByUserID(ctx context.Context, userID uuid.UUID) (*model.Client, error) {
	const op = "users.repository.postgres.GetClientByUserID"

	var c model.Client
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, first_name, last_name, phone, avatar_url, created_at, updated_at
		 FROM clients WHERE user_id = $1`,
		userID,
	).Scan(&c.ID, &c.UserID, &c.FirstName, &c.LastName, &c.Phone, &c.AvatarURL, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("[%s]: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return &c, nil
}

func (r *repository) UpdateClientAvatarURL(ctx context.Context, id uuid.UUID, objectName string) error {
	const op = "users.repository.postgres.UpdateClientAvatarURL"

	_, err := r.db.ExecContext(ctx,
		`UPDATE clients SET avatar_url = $1 WHERE id = $2`,
		objectName, id,
	)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func (r *repository) GetClientsByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Client, error) {
	const op = "users.repository.postgres.GetClientsByIDs"

	if len(ids) == 0 {
		return []model.Client{}, nil
	}

	query := `
		SELECT id, user_id, first_name, phone, avatar_url,  
			created_at, updated_at
		FROM clients
		WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("[%s]: query failed: %w", op, err)
	}
	defer rows.Close()

	clients := make([]model.Client, 0, len(ids))
	for rows.Next() {
		var client model.Client
		if err := rows.Scan(
			&client.ID, &client.UserID, &client.FirstName, &client.Phone, &client.AvatarURL,
			&client.CreatedAt, &client.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: scan failed: %w", op, err)
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[%s]: rows iteration failed: %w", op, err)
	}

	return clients, nil
}
