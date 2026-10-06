package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
)

// MerchantHandler groups merchant gRPC handlers.
type MerchantHandler interface {
	MerchantQueryHandleGrpc
}

type merchantHandler struct {
	MerchantQueryHandleGrpc
}

// MerchantDocumentHandler groups merchant document gRPC handlers.
type MerchantDocumentHandler interface {
	MerchantDocumentQueryHandleGrpc
}

type merchantDocumentHandler struct {
	MerchantDocumentQueryHandleGrpc
}

// NewHandler initializes merchant gRPC handlers.
func NewHandler(svc *service.Service) (MerchantHandler, MerchantDocumentHandler) {
	merchant := &merchantHandler{
		MerchantQueryHandleGrpc: NewMerchantHandleGrpc(svc.MerchantQuery, svc.MerchantCommand),
	}
	merchantDoc := &merchantDocumentHandler{
		MerchantDocumentQueryHandleGrpc: NewMerchantDocumentHandleGrpc(svc.MerchantDocumentQuery, svc.MerchantDocumentCommand),
	}
	return merchant, merchantDoc
}
