package users_transport_grpc

import (
	"context"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
)


type UsersService interface {
	GetUser(
		ctx context.Context,
		id int,
	) (domain.User,error)
	GetUsers(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.User,error)

}