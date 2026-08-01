package core_pgx_pool

import (
	"errors"
	"fmt"

	core_postgres_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxRows struct {
	pgx.Rows
}

type pgxRow struct {
	pgx.Row
}

type pgxCommandTag struct {
	pgconn.CommandTag
}


func (r pgxRow) Scan(dest ... any) error  {
	err := r.Row.Scan(dest...)

	if err != nil {
		return  mapErrors(err)

	}
	return nil
}



func mapErrors(err error) error {

		const (
		pgxViolatesForeingKeyErrorCode = "23503"
	)

	if errors.Is(err,pgx.ErrNoRows){
			return core_postgres_pool.ErrNoRows
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == pgxViolatesForeingKeyErrorCode {
			return fmt.Errorf(
				"%v: %w",
				err,
				core_postgres_pool.ErrViolatesForeingKey,
			)
		}
	}

	return fmt.Errorf(
		"%v:%w",
		err,
		core_postgres_pool.ErrUnknown,
	)
}