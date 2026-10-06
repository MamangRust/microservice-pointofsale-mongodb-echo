package handler

import (
	pb "github.com/MamangRust/microservice-point-of-sale-pb/user"
)

type UserQueryHandleGrpc interface {
	pb.UserServiceServer
}
