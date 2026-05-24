package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

func (r *repository) CreateReview(ctx context.Context, review model.Review) (*model.Review, error) {
	const op = "reviews.repository.postgres.CreateReview"

	query := `
		INSERT INTO reviews (id, client_id, master_id, appointment_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, client_id, master_id, appointment_id, rating, comment, created_at, updated_at
	`
	var rv model.Review
	err := r.db.QueryRowContext(ctx, query,
		review.ID, review.ClientID, review.MasterID, review.AppointmentID,
		review.Rating, review.Comment, review.CreatedAt, review.UpdatedAt,
	).Scan(
		&rv.ID, &rv.ClientID, &rv.MasterID, &rv.AppointmentID,
		&rv.Rating, &rv.Comment, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapReviewErrors(err))
	}
	return &rv, nil
}

func (r *repository) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	const op = "reviews.repository.postgres.GetReviewByID"

	query := `
		SELECT id, client_id, master_id, appointment_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE id = $1
	`
	var rv model.Review
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rv.ID, &rv.ClientID, &rv.MasterID, &rv.AppointmentID,
		&rv.Rating, &rv.Comment, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapReviewErrors(err))
	}
	return &rv, nil
}

func (r *repository) GetReviewByAppointmentID(ctx context.Context, appointmentID uuid.UUID) (*model.Review, error) {
	const op = "reviews.repository.postgres.GetReviewByAppointmentID"

	query := `
		SELECT id, client_id, master_id, appointment_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE appointment_id = $1
	`
	var rv model.Review
	err := r.db.QueryRowContext(ctx, query, appointmentID).Scan(
		&rv.ID, &rv.ClientID, &rv.MasterID, &rv.AppointmentID,
		&rv.Rating, &rv.Comment, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapReviewErrors(err))
	}
	return &rv, nil
}

func (r *repository) DeleteReview(ctx context.Context, id uuid.UUID) error {
	const op = "reviews.repository.postgres.DeleteReview"

	res, err := r.db.ExecContext(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, mapReviewErrors(err))
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	if n == 0 {
		return fmt.Errorf("[%s]: %w", op, mapReviewErrors(errNoRows))
	}
	return nil
}

func (r *repository) RecalcMasterRating(ctx context.Context, masterID uuid.UUID) error {
	const op = "reviews.repository.postgres.RecalcMasterRating"

	query := `
		UPDATE masters
		SET (rating, review_count) = (
			SELECT COALESCE(AVG(rating)::DECIMAL(3,2), 0), COUNT(*)
			FROM reviews
			WHERE master_id = $1
		)
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, masterID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	return nil
}

func (r *repository) GetReviewsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]model.Review, error) {
	const op = "reviews.repository.postgres.GetReviewsByClientID"

	query := `
		SELECT id, client_id, master_id, appointment_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapReviewErrors(err))
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(
			&rv.ID, &rv.ClientID, &rv.MasterID, &rv.AppointmentID,
			&rv.Rating, &rv.Comment, &rv.CreatedAt, &rv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: %w", op, err)
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}

func (r *repository) GetReviewsByMasterID(ctx context.Context, masterID uuid.UUID, limit, offset uint64) ([]model.Review, error) {
	const op = "reviews.repository.postgres.GetReviewsByMasterID"

	query := `
		SELECT id, client_id, master_id, appointment_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE master_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, masterID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapReviewErrors(err))
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(
			&rv.ID, &rv.ClientID, &rv.MasterID, &rv.AppointmentID,
			&rv.Rating, &rv.Comment, &rv.CreatedAt, &rv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("[%s]: %w", op, err)
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}
