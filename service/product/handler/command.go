package handler

import (
	"context"
	stderrors "errors"

	pb "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	product_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *productQueryHandleGrpc) Create(ctx context.Context, request *pb.CreateProductRequest) (*pb.ApiResponseProduct, error) {
	req := &requests.CreateProductRequest{
		MerchantID: int(request.GetMerchantId()), CategoryID: int(request.GetCategoryId()),
		Name: request.GetName(), Description: request.GetDescription(),
		Price: int(request.GetPrice()), CountInStock: int(request.GetCountInStock()),
		Brand: request.GetBrand(), Weight: int(request.GetWeight()), ImageProduct: request.GetImageProduct(),
	}
	if err := req.Validate(); err != nil {
		return nil, product_errors.ErrGrpcValidateCreateProduct
	}
	product, err := s.productCommandService.CreateProduct(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProduct{Status: "success", Message: "Successfully created product", Data: mapResponseProduct(product)}, nil
}

func (s *productQueryHandleGrpc) Update(ctx context.Context, request *pb.UpdateProductRequest) (*pb.ApiResponseProduct, error) {
	id := int(request.GetProductId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	req := &requests.UpdateProductRequest{
		ProductID: &id, MerchantID: int(request.GetMerchantId()), CategoryID: int(request.GetCategoryId()),
		Name: request.GetName(), Description: request.GetDescription(),
		Price: int(request.GetPrice()), CountInStock: int(request.GetCountInStock()),
		Brand: request.GetBrand(), Weight: int(request.GetWeight()), ImageProduct: request.GetImageProduct(),
	}
	if err := req.Validate(); err != nil {
		return nil, product_errors.ErrGrpcValidateUpdateProduct
	}
	product, err := s.productCommandService.UpdateProduct(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProduct{Status: "success", Message: "Successfully updated product", Data: mapResponseProduct(product)}, nil
}

func (s *productQueryHandleGrpc) DecrementStock(ctx context.Context, request *pb.AdjustProductStockRequest) (*pb.ApiResponseProduct, error) {
	return s.adjustStock(ctx, request, s.productCommandService.DecrementStock)
}

func (s *productQueryHandleGrpc) IncrementStock(ctx context.Context, request *pb.AdjustProductStockRequest) (*pb.ApiResponseProduct, error) {
	return s.adjustStock(ctx, request, s.productCommandService.IncrementStock)
}

func (s *productQueryHandleGrpc) adjustStock(
	ctx context.Context,
	request *pb.AdjustProductStockRequest,
	apply func(context.Context, int, int) (*models.Product, error),
) (*pb.ApiResponseProduct, error) {
	id := int(request.GetProductId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	if request.GetQuantity() <= 0 {
		return nil, product_errors.ErrGrpcInvalidQuantity
	}

	product, err := apply(ctx, id, int(request.GetQuantity()))
	if err != nil {
		if stderrors.Is(err, product_errors.ErrInsufficientProductStock) {
			return nil, product_errors.ErrGrpcInsufficientStock
		}
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProduct{Status: "success", Message: "Successfully adjusted product stock", Data: mapResponseProduct(product)}, nil
}

func (s *productQueryHandleGrpc) TrashedProduct(ctx context.Context, request *pb.FindByIdProductRequest) (*pb.ApiResponseProductDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	product, err := s.productCommandService.TrashProduct(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProductDeleteAt{Status: "success", Message: "Successfully trashed product", Data: mapResponseProductDeleteAt(product)}, nil
}

func (s *productQueryHandleGrpc) RestoreProduct(ctx context.Context, request *pb.FindByIdProductRequest) (*pb.ApiResponseProductDeleteAt, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	product, err := s.productCommandService.RestoreProduct(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProductDeleteAt{Status: "success", Message: "Successfully restored product", Data: mapResponseProductDeleteAt(product)}, nil
}

func (s *productQueryHandleGrpc) DeleteProductPermanent(ctx context.Context, request *pb.FindByIdProductRequest) (*pb.ApiResponseProductDelete, error) {
	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	_, err := s.productCommandService.DeleteProductPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProductDelete{Status: "success", Message: "Successfully deleted Product permanently"}, nil
}

func (s *productQueryHandleGrpc) RestoreAllProduct(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseProductAll, error) {
	_, err := s.productCommandService.RestoreAllProducts(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProductAll{Status: "success", Message: "Successfully restore all Product"}, nil
}

func (s *productQueryHandleGrpc) DeleteAllProductPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseProductAll, error) {
	_, err := s.productCommandService.DeleteAllProductsPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseProductAll{Status: "success", Message: "Successfully delete Product permanen"}, nil
}
