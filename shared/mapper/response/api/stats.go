package response_api

import (
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	orderpb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

// StatsResponseMapper maps the stats-reader (ClickHouse) gRPC responses to the
// REST shapes exposed by the API gateway (F4 §7.4). ByMerchant/ById variants
// reuse these mappers because their pb responses are identical.
type StatsResponseMapper interface {
	ToApiResponseOrderMonthlyTotalRevenue(res *orderpb.ApiResponseOrderMonthlyTotalRevenue) *response.ApiResponseOrderMonthlyTotalRevenue
	ToApiResponseOrderYearlyTotalRevenue(res *orderpb.ApiResponseOrderYearlyTotalRevenue) *response.ApiResponseOrderYearlyTotalRevenue
	ToApiResponseOrderMonthly(res *orderpb.ApiResponseOrderMonthly) *response.ApiResponseOrderMonthly
	ToApiResponseOrderYearly(res *orderpb.ApiResponseOrderYearly) *response.ApiResponseOrderYearly

	ToApiResponseCashierMonthlyTotalSales(res *cashierpb.ApiResponseCashierMonthlyTotalSales) *response.ApiResponseCashierMonthlyTotalSales
	ToApiResponseCashierYearlyTotalSales(res *cashierpb.ApiResponseCashierYearlyTotalSales) *response.ApiResponseCashierYearlyTotalSales
	ToApiResponseCashierMonthSales(res *cashierpb.ApiResponseCashierMonthSales) *response.ApiResponseCashierMonthSales
	ToApiResponseCashierYearSales(res *cashierpb.ApiResponseCashierYearSales) *response.ApiResponseCashierYearSales

	ToApiResponseCategoryMonthlyTotalPrice(res *categorypb.ApiResponseCategoryMonthlyTotalPrice) *response.ApiResponseCategoryMonthlyTotalPrice
	ToApiResponseCategoryYearlyTotalPrice(res *categorypb.ApiResponseCategoryYearlyTotalPrice) *response.ApiResponseCategoryYearlyTotalPrice
	ToApiResponseCategoryMonthPrice(res *categorypb.ApiResponseCategoryMonthPrice) *response.ApiResponseCategoryMonthPrice
	ToApiResponseCategoryYearPrice(res *categorypb.ApiResponseCategoryYearPrice) *response.ApiResponseCategoryYearPrice

	ToApiResponseTransactionMonthSuccess(res *transactionpb.ApiResponseTransactionMonthAmountSuccess) *response.ApiResponsesTransactionMonthSuccess
	ToApiResponseTransactionYearSuccess(res *transactionpb.ApiResponseTransactionYearAmountSuccess) *response.ApiResponsesTransactionYearSuccess
	ToApiResponseTransactionMonthFailed(res *transactionpb.ApiResponseTransactionMonthAmountFailed) *response.ApiResponsesTransactionMonthFailed
	ToApiResponseTransactionYearFailed(res *transactionpb.ApiResponseTransactionYearAmountFailed) *response.ApiResponsesTransactionYearFailed
	ToApiResponseTransactionMonthMethod(res *transactionpb.ApiResponseTransactionMonthPaymentMethod) *response.ApiResponsesTransactionMonthMethod
	ToApiResponseTransactionYearMethod(res *transactionpb.ApiResponseTransactionYearPaymentmethod) *response.ApiResponsesTransactionYearMethod
}

type statsResponseMapper struct{}

func NewStatsResponseMapper() *statsResponseMapper {
	return &statsResponseMapper{}
}

func (s *statsResponseMapper) ToApiResponseOrderMonthlyTotalRevenue(res *orderpb.ApiResponseOrderMonthlyTotalRevenue) *response.ApiResponseOrderMonthlyTotalRevenue {
	data := make([]*response.OrderMonthlyTotalRevenueResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderMonthlyTotalRevenueResponse{
			Year:           row.GetYear(),
			Month:          row.GetMonth(),
			OrderCount:     int(row.GetOrderCount()),
			TotalRevenue:   int(row.GetTotalRevenue()),
			TotalItemsSold: int(row.GetTotalItemsSold()),
		})
	}
	return &response.ApiResponseOrderMonthlyTotalRevenue{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderYearlyTotalRevenue(res *orderpb.ApiResponseOrderYearlyTotalRevenue) *response.ApiResponseOrderYearlyTotalRevenue {
	data := make([]*response.OrderYearlyTotalRevenueResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderYearlyTotalRevenueResponse{
			Year:               row.GetYear(),
			OrderCount:         int(row.GetOrderCount()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			TotalItemsSold:     int(row.GetTotalItemsSold()),
			ActiveCashiers:     int(row.GetActiveCashiers()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseOrderYearlyTotalRevenue{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderMonthly(res *orderpb.ApiResponseOrderMonthly) *response.ApiResponseOrderMonthly {
	data := make([]*response.OrderMonthlyResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderMonthlyResponse{
			Month:          row.GetMonth(),
			OrderCount:     int(row.GetOrderCount()),
			TotalRevenue:   int(row.GetTotalRevenue()),
			TotalItemsSold: int(row.GetTotalItemsSold()),
		})
	}
	return &response.ApiResponseOrderMonthly{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderYearly(res *orderpb.ApiResponseOrderYearly) *response.ApiResponseOrderYearly {
	data := make([]*response.OrderYearlyResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderYearlyResponse{
			Year:               row.GetYear(),
			OrderCount:         int(row.GetOrderCount()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			TotalItemsSold:     int(row.GetTotalItemsSold()),
			ActiveCashiers:     int(row.GetActiveCashiers()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseOrderYearly{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierMonthlyTotalSales(res *cashierpb.ApiResponseCashierMonthlyTotalSales) *response.ApiResponseCashierMonthlyTotalSales {
	data := make([]*response.CashierResponseMonthTotalSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseMonthTotalSales{
			Year:       row.GetYear(),
			Month:      row.GetMonth(),
			TotalSales: int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierMonthlyTotalSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierYearlyTotalSales(res *cashierpb.ApiResponseCashierYearlyTotalSales) *response.ApiResponseCashierYearlyTotalSales {
	data := make([]*response.CashierResponseYearTotalSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseYearTotalSales{
			Year:       row.GetYear(),
			TotalSales: int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierYearlyTotalSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierMonthSales(res *cashierpb.ApiResponseCashierMonthSales) *response.ApiResponseCashierMonthSales {
	data := make([]*response.CashierResponseMonthSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseMonthSales{
			Month:       row.GetMonth(),
			CashierID:   int(row.GetCashierId()),
			CashierName: row.GetCashierName(),
			OrderCount:  int(row.GetOrderCount()),
			TotalSales:  int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierMonthSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierYearSales(res *cashierpb.ApiResponseCashierYearSales) *response.ApiResponseCashierYearSales {
	data := make([]*response.CashierResponseYearSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseYearSales{
			Year:        row.GetYear(),
			CashierID:   int(row.GetCashierId()),
			CashierName: row.GetCashierName(),
			OrderCount:  int(row.GetOrderCount()),
			TotalSales:  int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierYearSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryMonthlyTotalPrice(res *categorypb.ApiResponseCategoryMonthlyTotalPrice) *response.ApiResponseCategoryMonthlyTotalPrice {
	data := make([]*response.CategoriesMonthlyTotalPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoriesMonthlyTotalPriceResponse{
			Year:         row.GetYear(),
			Month:        row.GetMonth(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryMonthlyTotalPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryYearlyTotalPrice(res *categorypb.ApiResponseCategoryYearlyTotalPrice) *response.ApiResponseCategoryYearlyTotalPrice {
	data := make([]*response.CategoriesYearlyTotalPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoriesYearlyTotalPriceResponse{
			Year:         row.GetYear(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryYearlyTotalPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryMonthPrice(res *categorypb.ApiResponseCategoryMonthPrice) *response.ApiResponseCategoryMonthPrice {
	data := make([]*response.CategoryMonthPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoryMonthPriceResponse{
			Month:        row.GetMonth(),
			CategoryID:   int(row.GetCategoryId()),
			CategoryName: row.GetCategoryName(),
			OrderCount:   int(row.GetOrderCount()),
			ItemsSold:    int(row.GetItemsSold()),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryMonthPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryYearPrice(res *categorypb.ApiResponseCategoryYearPrice) *response.ApiResponseCategoryYearPrice {
	data := make([]*response.CategoryYearPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoryYearPriceResponse{
			Year:               row.GetYear(),
			CategoryID:         int(row.GetCategoryId()),
			CategoryName:       row.GetCategoryName(),
			OrderCount:         int(row.GetOrderCount()),
			ItemsSold:          int(row.GetItemsSold()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseCategoryYearPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseTransactionMonthSuccess(res *transactionpb.ApiResponseTransactionMonthAmountSuccess) *response.ApiResponsesTransactionMonthSuccess {
	data := make([]*response.TransactionMonthlyAmountSuccessResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyAmountSuccessResponse{
			Year:         row.GetYear(),
			Month:        row.GetMonth(),
			TotalSuccess: int(row.GetTotalSuccess()),
			TotalAmount:  int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthSuccess{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseTransactionYearSuccess(res *transactionpb.ApiResponseTransactionYearAmountSuccess) *response.ApiResponsesTransactionYearSuccess {
	data := make([]*response.TransactionYearlyAmountSuccessResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyAmountSuccessResponse{
			Year:         row.GetYear(),
			TotalSuccess: int(row.GetTotalSuccess()),
			TotalAmount:  int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearSuccess{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseTransactionMonthFailed(res *transactionpb.ApiResponseTransactionMonthAmountFailed) *response.ApiResponsesTransactionMonthFailed {
	data := make([]*response.TransactionMonthlyAmountFailedResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyAmountFailedResponse{
			Year:        row.GetYear(),
			Month:       row.GetMonth(),
			TotalFailed: int(row.GetTotalFailed()),
			TotalAmount: int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthFailed{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseTransactionYearFailed(res *transactionpb.ApiResponseTransactionYearAmountFailed) *response.ApiResponsesTransactionYearFailed {
	data := make([]*response.TransactionYearlyAmountFailedResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyAmountFailedResponse{
			Year:        row.GetYear(),
			TotalFailed: int(row.GetTotalFailed()),
			TotalAmount: int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearFailed{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseTransactionMonthMethod(res *transactionpb.ApiResponseTransactionMonthPaymentMethod) *response.ApiResponsesTransactionMonthMethod {
	data := make([]*response.TransactionMonthlyMethodResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyMethodResponse{
			Month:             row.GetMonth(),
			PaymentMethod:     row.GetPaymentMethod(),
			TotalTransactions: int(row.GetTotalTransactions()),
			TotalAmount:       int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthMethod{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseTransactionYearMethod(res *transactionpb.ApiResponseTransactionYearPaymentmethod) *response.ApiResponsesTransactionYearMethod {
	data := make([]*response.TransactionYearlyMethodResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyMethodResponse{
			Year:              row.GetYear(),
			PaymentMethod:     row.GetPaymentMethod(),
			TotalTransactions: int(row.GetTotalTransactions()),
			TotalAmount:       int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearMethod{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}
