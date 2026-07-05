package core_pgx_pool

import (
	"errors"

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
		if errors.Is(err,pgx.ErrNoRows){
			return core_postgres_pool.ErrNoRows
		}
		return err 
	}

	return nil

}