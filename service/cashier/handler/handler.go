package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-cashier/service"
)

// Handler groups all cashier gRPC handlers.
type Handler interface {
	CashierQueryHandleGrpc
}

type handler struct {
	CashierQueryHandleGrpc
}

// NewHandler initializes cashier gRPC handlers.
func NewHandler(svc *service.Service) Handler {
	return &handler{
		CashierQueryHandleGrpc: NewCashierHandleGrpc(svc.CashierQuery, svc.CashierCommand),
	}
}
