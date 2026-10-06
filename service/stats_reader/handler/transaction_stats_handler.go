package handler

import (
	"context"
	"fmt"

	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-stats-reader/repository"
	"go.uber.org/zap"
)

// TransactionStatsHandler serves transaction aggregates from ClickHouse
// (transaction_daily): one struct registers both stats services — status
// (success vs failed) and payment method.
type TransactionStatsHandler struct {
	statspb.UnimplementedTransactionStatsStatusServiceServer
	statspb.UnimplementedTransactionStatsMethodServiceServer
	repo  repository.Repository
	cache *StatsCache
	log   logger.LoggerInterface
}

func NewTransactionStatsHandler(repo repository.Repository, cache *StatsCache, log logger.LoggerInterface) *TransactionStatsHandler {
	return &TransactionStatsHandler{repo: repo, cache: cache, log: log}
}

func (h *TransactionStatsHandler) FindMonthStatusSuccess(ctx context.Context, req *transactionpb.FindMonthlyTransactionStatus) (*transactionpb.ApiResponseTransactionMonthAmountSuccess, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-status-success:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthAmountSuccess](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthStatusSuccess(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthStatusSuccess failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthAmountSuccess{
		Status:  "success",
		Message: "Monthly successful transactions retrieved successfully",
		Data:    mapTransactionMonthAmountSuccess(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearStatusSuccess(ctx context.Context, req *transactionpb.FindYearlyTransactionStatus) (*transactionpb.ApiResponseTransactionYearAmountSuccess, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-status-success:%d", req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearAmountSuccess](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearStatusSuccess(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearStatusSuccess failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearAmountSuccess{
		Status:  "success",
		Message: "Yearly successful transactions retrieved successfully",
		Data:    mapTransactionYearAmountSuccess(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthStatusFailed(ctx context.Context, req *transactionpb.FindMonthlyTransactionStatus) (*transactionpb.ApiResponseTransactionMonthAmountFailed, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-status-failed:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthAmountFailed](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthStatusFailed(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthStatusFailed failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthAmountFailed{
		Status:  "success",
		Message: "Monthly failed transactions retrieved successfully",
		Data:    mapTransactionMonthAmountFailed(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearStatusFailed(ctx context.Context, req *transactionpb.FindYearlyTransactionStatus) (*transactionpb.ApiResponseTransactionYearAmountFailed, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-status-failed:%d", req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearAmountFailed](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearStatusFailed(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearStatusFailed failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearAmountFailed{
		Status:  "success",
		Message: "Yearly failed transactions retrieved successfully",
		Data:    mapTransactionYearAmountFailed(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthStatusSuccessByMerchant(ctx context.Context, req *transactionpb.FindMonthlyTransactionStatusByMerchant) (*transactionpb.ApiResponseTransactionMonthAmountSuccess, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-status-success:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthAmountSuccess](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthStatusSuccessByMerchant(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthStatusSuccessByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthAmountSuccess{
		Status:  "success",
		Message: "Monthly successful transactions by merchant retrieved successfully",
		Data:    mapTransactionMonthAmountSuccess(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearStatusSuccessByMerchant(ctx context.Context, req *transactionpb.FindYearlyTransactionStatusByMerchant) (*transactionpb.ApiResponseTransactionYearAmountSuccess, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-status-success:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearAmountSuccess](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearStatusSuccessByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearStatusSuccessByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearAmountSuccess{
		Status:  "success",
		Message: "Yearly successful transactions by merchant retrieved successfully",
		Data:    mapTransactionYearAmountSuccess(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthStatusFailedByMerchant(ctx context.Context, req *transactionpb.FindMonthlyTransactionStatusByMerchant) (*transactionpb.ApiResponseTransactionMonthAmountFailed, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-status-failed:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthAmountFailed](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthStatusFailedByMerchant(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthStatusFailedByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthAmountFailed{
		Status:  "success",
		Message: "Monthly failed transactions by merchant retrieved successfully",
		Data:    mapTransactionMonthAmountFailed(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearStatusFailedByMerchant(ctx context.Context, req *transactionpb.FindYearlyTransactionStatusByMerchant) (*transactionpb.ApiResponseTransactionYearAmountFailed, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-status-failed:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearAmountFailed](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearStatusFailedByMerchant(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearStatusFailedByMerchant failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearAmountFailed{
		Status:  "success",
		Message: "Yearly failed transactions by merchant retrieved successfully",
		Data:    mapTransactionYearAmountFailed(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthMethodSuccess(ctx context.Context, req *transactionpb.MonthTransactionMethod) (*transactionpb.ApiResponseTransactionMonthPaymentMethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-method-success:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthPaymentMethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthMethodSuccess(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthMethodSuccess failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Monthly successful payment methods retrieved successfully",
		Data:    mapTransactionMonthMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearMethodSuccess(ctx context.Context, req *transactionpb.YearTransactionMethod) (*transactionpb.ApiResponseTransactionYearPaymentmethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-method-success:%d", req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearPaymentmethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearMethodSuccess(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearMethodSuccess failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Yearly successful payment methods retrieved successfully",
		Data:    mapTransactionYearMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthMethodFailed(ctx context.Context, req *transactionpb.MonthTransactionMethod) (*transactionpb.ApiResponseTransactionMonthPaymentMethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-method-failed:%d:%d", req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthPaymentMethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthMethodFailed(ctx, int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		h.log.Error("FindMonthMethodFailed failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Monthly failed payment methods retrieved successfully",
		Data:    mapTransactionMonthMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearMethodFailed(ctx context.Context, req *transactionpb.YearTransactionMethod) (*transactionpb.ApiResponseTransactionYearPaymentmethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-method-failed:%d", req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearPaymentmethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearMethodFailed(ctx, int(req.GetYear()))
	if err != nil {
		h.log.Error("FindYearMethodFailed failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Yearly failed payment methods retrieved successfully",
		Data:    mapTransactionYearMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthMethodByMerchantSuccess(ctx context.Context, req *transactionpb.MonthTransactionMethodByMerchant) (*transactionpb.ApiResponseTransactionMonthPaymentMethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-method-success:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthPaymentMethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthMethodByMerchantSuccess(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthMethodByMerchantSuccess failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Monthly successful payment methods by merchant retrieved successfully",
		Data:    mapTransactionMonthMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearMethodByMerchantSuccess(ctx context.Context, req *transactionpb.YearTransactionMethodByMerchant) (*transactionpb.ApiResponseTransactionYearPaymentmethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-method-success:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearPaymentmethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearMethodByMerchantSuccess(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearMethodByMerchantSuccess failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Yearly successful payment methods by merchant retrieved successfully",
		Data:    mapTransactionYearMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindMonthMethodByMerchantFailed(ctx context.Context, req *transactionpb.MonthTransactionMethodByMerchant) (*transactionpb.ApiResponseTransactionMonthPaymentMethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:month-method-failed:merchant:%d:%d:%d", req.GetMerchantId(), req.GetYear(), req.GetMonth())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionMonthPaymentMethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetMonthMethodByMerchantFailed(ctx, int(req.GetYear()), int(req.GetMonth()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindMonthMethodByMerchantFailed failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Monthly failed payment methods by merchant retrieved successfully",
		Data:    mapTransactionMonthMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

func (h *TransactionStatsHandler) FindYearMethodByMerchantFailed(ctx context.Context, req *transactionpb.YearTransactionMethodByMerchant) (*transactionpb.ApiResponseTransactionYearPaymentmethod, error) {
	key := fmt.Sprintf("stats:reader:transaction:year-method-failed:merchant:%d:%d", req.GetMerchantId(), req.GetYear())
	if cached, found := CacheGet[transactionpb.ApiResponseTransactionYearPaymentmethod](ctx, h.cache, key); found {
		return cached, nil
	}

	data, err := h.repo.GetYearMethodByMerchantFailed(ctx, int(req.GetYear()), int(req.GetMerchantId()))
	if err != nil {
		h.log.Error("FindYearMethodByMerchantFailed failed", zap.Error(err))
		return nil, err
	}

	resp := &transactionpb.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Yearly failed payment methods by merchant retrieved successfully",
		Data:    mapTransactionYearMethod(data),
	}
	CacheSet(ctx, h.cache, key, resp)
	return resp, nil
}

// --- Mappers ---

func mapTransactionMonthAmountSuccess(data []repository.TransactionMonthAmount) []*transactionpb.TransactionMonthlyAmountSuccess {
	var out []*transactionpb.TransactionMonthlyAmountSuccess
	for _, d := range data {
		out = append(out, &transactionpb.TransactionMonthlyAmountSuccess{
			Year:         d.Year,
			Month:        d.Month,
			TotalSuccess: int32(d.TotalCount),
			TotalAmount:  int32(d.TotalAmount),
		})
	}
	return out
}

func mapTransactionYearAmountSuccess(data []repository.TransactionYearAmount) []*transactionpb.TransactionYearlyAmountSuccess {
	var out []*transactionpb.TransactionYearlyAmountSuccess
	for _, d := range data {
		out = append(out, &transactionpb.TransactionYearlyAmountSuccess{
			Year:         d.Year,
			TotalSuccess: int32(d.TotalCount),
			TotalAmount:  int32(d.TotalAmount),
		})
	}
	return out
}

func mapTransactionMonthAmountFailed(data []repository.TransactionMonthAmount) []*transactionpb.TransactionMonthlyAmountFailed {
	var out []*transactionpb.TransactionMonthlyAmountFailed
	for _, d := range data {
		out = append(out, &transactionpb.TransactionMonthlyAmountFailed{
			Year:        d.Year,
			Month:       d.Month,
			TotalFailed: int32(d.TotalCount),
			TotalAmount: int32(d.TotalAmount),
		})
	}
	return out
}

func mapTransactionYearAmountFailed(data []repository.TransactionYearAmount) []*transactionpb.TransactionYearlyAmountFailed {
	var out []*transactionpb.TransactionYearlyAmountFailed
	for _, d := range data {
		out = append(out, &transactionpb.TransactionYearlyAmountFailed{
			Year:        d.Year,
			TotalFailed: int32(d.TotalCount),
			TotalAmount: int32(d.TotalAmount),
		})
	}
	return out
}

func mapTransactionMonthMethod(data []repository.TransactionMonthMethod) []*transactionpb.TransactionMonthlyMethod {
	var out []*transactionpb.TransactionMonthlyMethod
	for _, d := range data {
		out = append(out, &transactionpb.TransactionMonthlyMethod{
			Month:             d.Month,
			PaymentMethod:     d.PaymentMethod,
			TotalTransactions: int32(d.TotalTransactions),
			TotalAmount:       int32(d.TotalAmount),
		})
	}
	return out
}

func mapTransactionYearMethod(data []repository.TransactionYearMethod) []*transactionpb.TransactionYearlyMethod {
	var out []*transactionpb.TransactionYearlyMethod
	for _, d := range data {
		out = append(out, &transactionpb.TransactionYearlyMethod{
			Year:              d.Year,
			PaymentMethod:     d.PaymentMethod,
			TotalTransactions: int32(d.TotalTransactions),
			TotalAmount:       int32(d.TotalAmount),
		})
	}
	return out
}
