package users_transport_grpc

import (
	"context"

	pb "github.com/Trykach34rus/Golang-todoapp/internal/features/users/proto"
)

func (s *UsersGRPCServer) GetUsers(
	ctx context.Context,
	req *pb.GetUsersRequest,
) (*pb.GetUsersResponse,error) {

  limit := int(req.Limit)
  offset := int(req.Offset)

	users,err := s.usersService.GetUsers(
		ctx,
		&limit,
		&offset,
	)


	if err != nil {
		return nil,err
	}


	return &pb.GetUsersResponse{
		Users: usersProtoFromDomains(users),
	},nil
}

