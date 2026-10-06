package repository

import (
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	categoryadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/category"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	ProductQuery   ProductQueryRepository
	ProductCommand ProductCommandRepository
	CategoryQuery  categoryadapter.Repository
	MerchantQuery  merchantadapter.Repository
}

type GuardOptions struct {
	Category []adapter.GuardOption
	Merchant []adapter.GuardOption
}

func NewRepositories(
	DB *mongo.Database,
	categoryClient pbcategory.CategoryServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	return &Repositories{
		ProductQuery:   NewProductQueryRepository(DB),
		ProductCommand: NewProductCommandRepository(DB),
		CategoryQuery:  categoryadapter.New(categoryClient, g.Category...),
		MerchantQuery:  merchantadapter.New(merchantClient, g.Merchant...),
	}
}
