package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// Local result types for cashier queries

type CashierResult struct {
	CashierID  int32
	MerchantID int32
	UserID     int32
	Name       string
	CreatedAt  *string
	UpdatedAt  *string
	TotalCount int64
}

type CashierResultDeleteAt struct {
	CashierID  int32
	MerchantID int32
	UserID     int32
	Name       string
	CreatedAt  *string
	UpdatedAt  *string
	DeletedAt  *string
	TotalCount int64
}

type MonthlyTotalSalesResult struct {
	Year       string
	Month      string
	TotalSales int64
}

type YearlyTotalSalesResult struct {
	Year       string
	TotalSales int64
}

type MonthlyCashierResult struct {
	Month       string
	CashierID   int32
	CashierName string
	OrderCount  int64
	TotalSales  int64
}

type YearlyCashierResult struct {
	Year        string
	CashierID   int32
	CashierName string
	OrderCount  int64
	TotalSales  int64
}

type CashierQueryRepository interface {
	FindAllCashiers(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResult, *int, error)
	FindById(ctx context.Context, cashier_id int) (*models.Cashier, error)
	FindByActive(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResultDeleteAt, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*CashierResult, *int, error)
}

type CashierCommandRepository interface {
	CreateCashier(ctx context.Context, request *requests.CreateCashierRequest) (*models.Cashier, error)
	UpdateCashier(ctx context.Context, request *requests.UpdateCashierRequest) (*models.Cashier, error)
	TrashedCashier(ctx context.Context, cashier_id int) (*models.Cashier, error)
	RestoreCashier(ctx context.Context, cashier_id int) (*models.Cashier, error)
	DeleteCashierPermanent(ctx context.Context, cashier_id int) (bool, error)
	RestoreAllCashier(ctx context.Context) (bool, error)
	DeleteAllCashierPermanent(ctx context.Context) (bool, error)
}
