package repository

import (
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	UserQuery      *useradapter.Repository
	MerchantQuery  merchantadapter.Repository
	CashierQuery   CashierQueryRepository
	CashierCommand CashierCommandRepository
}

// GuardOptions collects the dependency guards for each remote dependency.
type GuardOptions struct {
	User     []adapter.GuardOption
	Merchant []adapter.GuardOption
}

func NewRepositories(
	DB *mongo.Database,
	userClient pbuser.UserServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		UserQuery:      useradapter.New(userClient, g.User...),
		MerchantQuery:  merchantadapter.New(merchantClient, g.Merchant...),
		CashierQuery:   NewCashierQueryRepository(DB),
		CashierCommand: NewCashierCommandRepository(DB),
	}
}
