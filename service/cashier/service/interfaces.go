package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CashierQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResult, *int, error)
	FindById(ctx context.Context, cashierID int) (*models.Cashier, error)
	FindByActive(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResultDeleteAt, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResultDeleteAt, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*repository.CashierResult, *int, error)
}

type CashierCommandService interface {
	CreateCashier(ctx context.Context, req *requests.CreateCashierRequest) (*models.Cashier, error)
	UpdateCashier(ctx context.Context, req *requests.UpdateCashierRequest) (*models.Cashier, error)
	TrashedCashier(ctx context.Context, cashierID int) (*models.Cashier, error)
	RestoreCashier(ctx context.Context, cashierID int) (*models.Cashier, error)
	DeleteCashierPermanent(ctx context.Context, cashierID int) (bool, error)
	RestoreAllCashier(ctx context.Context) (bool, error)
	DeleteAllCashierPermanent(ctx context.Context) (bool, error)
}
