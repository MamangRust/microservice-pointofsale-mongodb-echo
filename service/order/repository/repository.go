package repository

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	OrderQuery       OrderQueryRepository
	OrderCommand     OrderCommandRepository
	CashierQuery     cashieradapter.QueryRepository
	MerchantQuery    merchantadapter.Repository
	ProductQuery     productadapter.QueryRepository
	ProductCommand   productadapter.CommandRepository
	OrderItemQuery   orderitemadapter.QueryRepository
	OrderItemCommand orderitemadapter.CommandRepository
}

type GuardOptions struct {
	Cashier   []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Product   []adapter.GuardOption
	OrderItem []adapter.GuardOption
}

func NewRepositories(
	db *mongo.Database,
	cashierClient pbcashier.CashierServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	productClient pbproduct.ProductServiceClient,
	orderItemClient pborderitem.OrderItemServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	productRepo := productadapter.New(productClient, g.Product...)
	orderItemRepo := orderitemadapter.New(orderItemClient, g.OrderItem...)
	return &Repositories{
		OrderQuery:       NewOrderQueryRepository(db),
		OrderCommand:     NewOrderCommandRepository(db),
		CashierQuery:     cashieradapter.New(cashierClient, g.Cashier...),
		MerchantQuery:    merchantadapter.New(merchantClient, g.Merchant...),
		ProductQuery:     productRepo,
		ProductCommand:   productRepo,
		OrderItemQuery:   orderItemRepo,
		OrderItemCommand: orderItemRepo,
	}
}
