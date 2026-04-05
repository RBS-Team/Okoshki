package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound     = errors.New("appointment not found")
	ErrTimeConflict = errors.New("time slot is already booked")
	ErrInternal     = errors.New("internal db error")
)

func mapAppointmentErrors(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23P01 - exclusion_violation (EXCLUDE констрейнт на пересечение времени)
		if pgErr.Code == "23P01" {
			return ErrTimeConflict
		}
	}

	return err
}
