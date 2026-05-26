package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RBS-Team/Okoshki/internal/domain"
)

var errNoRows = sql.ErrNoRows

func mapReviewErrors(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return domain.ErrConflict
		case "23503": // foreign_key_violation
			return domain.ErrNotFound
		}
	}
	return err
}
