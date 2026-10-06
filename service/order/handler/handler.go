package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-order/service"
)

// Handler groups all order gRPC handlers.
type Handler interface {
	OrderQueryHandleGrpc
}

type handler struct {
	OrderQueryHandleGrpc
}

// NewHandler initializes order gRPC handlers.
func NewHandler(svc *service.Service) Handler {
	return &handler{
		OrderQueryHandleGrpc: NewOrderHandleGrpc(svc.OrderQuery, svc.OrderCommand),
	}
}
