package handler

import (
	pb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
)

type MerchantQueryHandleGrpc interface {
	pb.MerchantServiceServer
}

type MerchantDocumentQueryHandleGrpc interface {
	pbmerchantdoc.MerchantDocumentServiceServer
}
