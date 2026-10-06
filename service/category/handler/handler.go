package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-category/service"
)

// Handler groups all category gRPC handlers.
type Handler interface {
	CategoryQueryHandleGrpc
}

type handler struct {
	CategoryQueryHandleGrpc
}

// NewHandler initializes category gRPC handlers.
func NewHandler(svc *service.Service) Handler {
	return &handler{
		CategoryQueryHandleGrpc: NewCategoryHandleGrpc(svc.CategoryQuery, svc.CategoryCommand),
	}
}
