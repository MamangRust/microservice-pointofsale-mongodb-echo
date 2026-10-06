package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-transacton/service"
)

// Handler groups all transaction gRPC handlers.
type Handler interface {
	TransactionQueryHandleGrpc
}

type handler struct {
	TransactionQueryHandleGrpc
}

// NewHandler initializes transaction gRPC handlers.
func NewHandler(svc *service.Service, logger logger.LoggerInterface) Handler {
	return &handler{
		TransactionQueryHandleGrpc: NewTransactionQueryHandleGrpc(svc, logger),
	}
}
