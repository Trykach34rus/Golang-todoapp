package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_errors "github.com/Trykach34rus/Golang-todoapp/internal/core/errors"
	core_postgres_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool"
)

func (r *UserRepository)PatchUser(
	ctx context.Context,
	id int,
	user domain.User,
) (domain.User, error) {

	ctx,cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE todoapp.users
	SET 
	  full_name=$1,
		phone_number=$2,
		version=version+1
	WHERE id=$3	AND version=$4
	RETURNING
	  id,
		version,
		full_name,
		phone_number;
		`

	row := r.pool.QueryRow(
		ctx,
		query,
		user.FullName,
		user.PhoneNumber,
		id,
		user.Version,
	)
	
	var userModel UserModel

	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.FullName,
		&userModel.PhoneNumber,
	)
	if err != nil {
		if errors.Is(err,core_postgres_pool.ErrNoRows) {
			return domain.User{},fmt.Errorf(
				"user with id='%d' concurrently accssesd: %w",
				id,
				core_errors.ErrConflict,
			)
		}

		return domain.User{},fmt.Errorf("scan error:%w",err)
	}

	userDomain := domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.FullName,
		user.PhoneNumber,
	)

	return userDomain,nil

}