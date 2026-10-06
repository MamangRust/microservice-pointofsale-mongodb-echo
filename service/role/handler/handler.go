package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-role/service"
)

// Handler groups all role gRPC handlers.
type Handler interface {
	RoleQueryHandleGrpc
	UserRoleHandleGrpc
}

type handler struct {
	RoleQueryHandleGrpc
	UserRoleHandleGrpc
}

// NewHandler initializes role gRPC handlers.
func NewHandler(svc *service.Service) Handler {
	return &handler{
		RoleQueryHandleGrpc: NewRoleHandleGrpc(svc.RoleQuery, svc.RoleCommand),
		UserRoleHandleGrpc:  NewUserRoleHandleGrpc(svc.RoleQuery, svc.RoleCommand),
	}
}
