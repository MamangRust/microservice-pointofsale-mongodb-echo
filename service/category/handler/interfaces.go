package handler

import (
	pb "github.com/MamangRust/microservice-point-of-sale-pb/category"
)

type CategoryQueryHandleGrpc interface {
	pb.CategoryServiceServer
}
