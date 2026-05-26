package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RBS-Team/Okoshki/internal/domain"
)

func mapAppointmentErrors(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01":
			return domain.ErrTimeConflict
		case "23503":
			if pgErr.ConstraintName == "appointments_client_id_fkey" {
				return domain.ErrUnauthorized
			}
		}
	}

	return err
}
