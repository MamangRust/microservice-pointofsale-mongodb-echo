package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-user/service"
)

// Handler groups all user gRPC handlers.
type Handler interface {
	UserQueryHandleGrpc
}

type handler struct {
	UserQueryHandleGrpc
}

// NewHandler initializes user gRPC handlers.
func NewHandler(svc *service.Service) Handler {
	return &handler{
		UserQueryHandleGrpc: NewUserHandleGrpc(svc.UserQuery, svc.UserCommand),
	}
}
