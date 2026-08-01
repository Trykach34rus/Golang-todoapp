package users_transport_grpc

import (
	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	pb "github.com/Trykach34rus/Golang-todoapp/internal/features/users/proto"
)


func userProtoFromDomain(user domain.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:          int32(user.ID),
		Version:     int32(user.Version),
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
	}
}

func usersProtoFromDomains(users []domain.User) []*pb.UserResponse {

	usersProto := make([]*pb.UserResponse,len(users))

	for i,user := range users {
		usersProto[i] = userProtoFromDomain(user)
	}

	return usersProto
}