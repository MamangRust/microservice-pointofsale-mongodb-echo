package handler

import (
	pb "github.com/MamangRust/microservice-point-of-sale-pb/order"
)

type OrderQueryHandleGrpc interface {
	pb.OrderServiceServer
}
