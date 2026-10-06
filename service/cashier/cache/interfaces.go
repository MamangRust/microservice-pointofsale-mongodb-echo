package mencache

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CashierQueryCache interface {
	GetCachedCashiersCache(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResult, *int, bool)
	SetCachedCashiersCache(ctx context.Context, req *requests.FindAllCashiers, res []*repository.CashierResult, total *int)

	GetCachedCashier(ctx context.Context, cashierID int) (*models.Cashier, bool)
	SetCachedCashier(ctx context.Context, res *models.Cashier)

	GetCachedCashiersActive(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResultDeleteAt, *int, bool)
	SetCachedCashiersActive(ctx context.Context, req *requests.FindAllCashiers, res []*repository.CashierResultDeleteAt, total *int)

	GetCachedCashiersTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*repository.CashierResultDeleteAt, *int, bool)
	SetCachedCashiersTrashed(ctx context.Context, req *requests.FindAllCashiers, res []*repository.CashierResultDeleteAt, total *int)

	GetCachedCashiersByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*repository.CashierResult, *int, bool)
	SetCachedCashiersByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant, res []*repository.CashierResult, total *int)
}

type CashierCommandCache interface {
	DeleteCashierCache(ctx context.Context, id int)
	DeleteCashierListCache(ctx context.Context)
}
