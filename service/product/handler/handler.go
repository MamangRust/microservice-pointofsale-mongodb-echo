package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-product/service"
)

// Handler groups all product gRPC handlers.
type Handler interface {
	ProductQueryHandleGrpc
}

type handler struct {
	ProductQueryHandleGrpc
}

// NewHandler initializes product gRPC handlers.
func NewHandler(svc *service.Service) Handler {
	return &handler{
		ProductQueryHandleGrpc: NewProductHandleGrpc(svc.ProductQuery, svc.ProductCommand),
	}
}
