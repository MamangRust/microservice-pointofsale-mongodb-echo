package handler

import (
	"context"
	"math"

	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/convert"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	transaction_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	"github.com/MamangRust/microservice-point-of-sale-transacton/service"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transactionQueryHandleGrpc struct {
	pbtransaction.UnimplementedTransactionServiceServer
	transactionQuery   service.TransactionQueryService
	transactionCommand service.TransactionCommandService
	logger             logger.LoggerInterface
}

func NewTransactionQueryHandleGrpc(
	svc *service.Service,
	logger logger.LoggerInterface,
) pbtransaction.TransactionServiceServer {
	return &transactionQueryHandleGrpc{
		transactionQuery:   svc.TransactionQuery,
		transactionCommand: svc.TransactionCommand,
		logger:             logger,
	}
}

func (s *transactionQueryHandleGrpc) FindAll(ctx context.Context, request *pbtransaction.FindAllTransactionRequest) (*pbtransaction.ApiResponsePaginationTransaction, error) {
	s.logger.Info("FindAll transactions called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransaction{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindAllTransactions(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll transactions success")

	return &pbtransaction.ApiResponsePaginationTransaction{
		Status:     "success",
		Message:    "Successfully fetched transaction",
		Data:       mapTxResponsesTransaction(transaction),
		Pagination: mapTxPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionQueryHandleGrpc) FindByMerchant(ctx context.Context, request *pbtransaction.FindAllTransactionMerchantRequest) (*pbtransaction.ApiResponsePaginationTransaction, error) {
	s.logger.Info("FindByMerchant transactions called", zap.Int32("merchantId", request.GetMerchantId()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchantID := int(request.GetMerchantId())

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransactionByMerchant{
		MerchantID: merchantID,
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindByMerchant(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByMerchant transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByMerchant transactions success")

	return &pbtransaction.ApiResponsePaginationTransaction{
		Status:     "success",
		Message:    "Successfully fetched transaction",
		Data:       mapTxResponsesTransactionByMerchant(transaction),
		Pagination: mapTxPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionQueryHandleGrpc) FindById(ctx context.Context, request *pbtransaction.FindByIdTransactionRequest) (*pbtransaction.ApiResponseTransaction, error) {
	s.logger.Info("FindById transaction called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, transaction_errors.ErrGrpcInvalidID
	}

	transaction, err := s.transactionQuery.FindById(ctx, id)
	if err != nil {
		s.logger.Error("FindById transaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById transaction success")

	return &pbtransaction.ApiResponseTransaction{
		Status:  "success",
		Message: "Successfully fetched transaction",
		Data:    mapTxResponseTransaction(transaction),
	}, nil
}

func (s *transactionQueryHandleGrpc) FindByActive(ctx context.Context, request *pbtransaction.FindAllTransactionRequest) (*pbtransaction.ApiResponsePaginationTransactionDeleteAt, error) {
	s.logger.Info("FindByActive transactions called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransaction{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive transactions success")

	return &pbtransaction.ApiResponsePaginationTransactionDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active transaction",
		Data:       mapTxResponsesTransactionActive(transaction),
		Pagination: mapTxPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pbtransaction.FindAllTransactionRequest) (*pbtransaction.ApiResponsePaginationTransactionDeleteAt, error) {
	s.logger.Info("FindByTrashed transactions called")

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransaction{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed transactions success")

	return &pbtransaction.ApiResponsePaginationTransactionDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed transaction",
		Data:       mapTxResponsesTransactionTrashed(transaction),
		Pagination: mapTxPaginationMeta(paginationMeta),
	}, nil
}

// Map helpers

func mapTxPaginationMeta(meta *pbcommon.PaginationMeta) *pbcommon.PaginationMeta {
	if meta == nil {
		return nil
	}
	return &pbcommon.PaginationMeta{
		CurrentPage:  meta.CurrentPage,
		PageSize:     meta.PageSize,
		TotalPages:   meta.TotalPages,
		TotalRecords: meta.TotalRecords,
	}
}

func mapTxResponseTransaction(transaction *models.Transaction) *pbtransaction.TransactionResponse {
	if transaction == nil {
		return nil
	}
	var changeAmount int32
	if transaction.ChangeAmount != nil {
		changeAmount = *transaction.ChangeAmount
	}
	var paymentStatus string
	if transaction.PaymentStatus != nil {
		paymentStatus = *transaction.PaymentStatus
	}
	return &pbtransaction.TransactionResponse{
		Id:            transaction.TransactionID,
		OrderId:       transaction.OrderID,
		MerchantId:    transaction.MerchantID,
		PaymentMethod: transaction.PaymentMethod,
		Amount:        transaction.Amount,
		ChangeAmount:  changeAmount,
		PaymentStatus: paymentStatus,
		CreatedAt:     convert.FormatTimePtr(transaction.CreatedAt),
		UpdatedAt:     convert.FormatTimePtr(transaction.UpdatedAt),
	}
}

func mapTxResponsesTransaction(transactions []*repository.TransactionResult) []*pbtransaction.TransactionResponse {
	var mappedTransactions []*pbtransaction.TransactionResponse
	for _, t := range transactions {
		if t == nil {
			continue
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponse{
			Id:            t.TransactionID,
			OrderId:       t.OrderID,
			MerchantId:    t.MerchantID,
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     t.CreatedAt,
			UpdatedAt:     t.CreatedAt,
		})
	}
	return mappedTransactions
}

func mapTxResponsesTransactionByMerchant(transactions []*repository.TransactionByMerchantResult) []*pbtransaction.TransactionResponse {
	var mappedTransactions []*pbtransaction.TransactionResponse
	for _, t := range transactions {
		if t == nil {
			continue
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponse{
			Id:            t.TransactionID,
			OrderId:       t.OrderID,
			MerchantId:    t.MerchantID,
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     t.CreatedAt,
			UpdatedAt:     t.CreatedAt,
		})
	}
	return mappedTransactions
}

func mapTxResponseTransactionDeleteAt(transaction *models.Transaction) *pbtransaction.TransactionResponseDeleteAt {
	if transaction == nil {
		return nil
	}
	var changeAmount int32
	if transaction.ChangeAmount != nil {
		changeAmount = *transaction.ChangeAmount
	}
	var paymentStatus string
	if transaction.PaymentStatus != nil {
		paymentStatus = *transaction.PaymentStatus
	}
	return &pbtransaction.TransactionResponseDeleteAt{
		Id:            transaction.TransactionID,
		OrderId:       transaction.OrderID,
		MerchantId:    transaction.MerchantID,
		PaymentMethod: transaction.PaymentMethod,
		Amount:        transaction.Amount,
		ChangeAmount:  changeAmount,
		PaymentStatus: paymentStatus,
		CreatedAt:     convert.FormatTimePtr(transaction.CreatedAt),
		UpdatedAt:     convert.FormatTimePtr(transaction.UpdatedAt),
		DeletedAt:     convert.TimeToWrappers(transaction.DeletedAt),
	}
}

func mapTxResponsesTransactionActive(transactions []*repository.TransactionResultDeleteAt) []*pbtransaction.TransactionResponseDeleteAt {
	var mappedTransactions []*pbtransaction.TransactionResponseDeleteAt
	for _, t := range transactions {
		if t == nil {
			continue
		}
		var deletedAt *wrapperspb.StringValue
		if t.DeletedAt != "" {
			deletedAt = wrapperspb.String(t.DeletedAt)
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponseDeleteAt{
			Id:            t.TransactionID,
			OrderId:       t.OrderID,
			MerchantId:    t.MerchantID,
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     t.CreatedAt,
			UpdatedAt:     t.CreatedAt,
			DeletedAt:     deletedAt,
		})
	}
	return mappedTransactions
}

func mapTxResponsesTransactionTrashed(transactions []*repository.TransactionResultDeleteAt) []*pbtransaction.TransactionResponseDeleteAt {
	return mapTxResponsesTransactionActive(transactions)
}
