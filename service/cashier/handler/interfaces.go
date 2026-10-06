package handler

import (
	pb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
)

type CashierQueryHandleGrpc interface {
	pb.CashierServiceServer
}
