package users_transport_grpc

import (
	"context"

	pb "github.com/Trykach34rus/Golang-todoapp/internal/features/users/proto"
)

func (s *UsersGRPCServer) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {

	user, err := s.usersService.GetUser(
		ctx,
		int(req.Id),
	)

	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{
		User: userProtoFromDomain(user),
	}, nil
}