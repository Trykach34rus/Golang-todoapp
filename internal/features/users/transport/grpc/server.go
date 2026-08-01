package users_transport_grpc

import (
	pb "github.com/Trykach34rus/Golang-todoapp/internal/features/users/proto"
)

type UsersGRPCServer struct {
	pb.UnimplementedUsersServiceServer

	usersService UsersService
}


func NewUsersGRPCServer(
	usersService UsersService,
) *UsersGRPCServer {

	return &UsersGRPCServer{
		usersService: usersService,
	}
}