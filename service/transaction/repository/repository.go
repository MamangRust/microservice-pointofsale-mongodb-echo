package repository

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	CashierQuery                 cashieradapter.QueryRepository
	MerchantQuery                merchantadapter.Repository
	OrderQuery                   orderadapter.QueryRepository
	OrderItemQuery               orderitemadapter.QueryRepository
	TransactionCommandRepository TransactionCommandRepository
	TransactionQueryRepository   TransactionQueryRepository
}

type GuardOptions struct {
	Cashier   []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Order     []adapter.GuardOption
	OrderItem []adapter.GuardOption
}

func NewRepositories(
	db *mongo.Database,
	cashierClient pbcashier.CashierServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	orderClient pborder.OrderServiceClient,
	orderItemClient pborderitem.OrderItemServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	return &Repositories{
		CashierQuery:                 cashieradapter.New(cashierClient, g.Cashier...),
		MerchantQuery:                merchantadapter.New(merchantClient, g.Merchant...),
		OrderQuery:                   orderadapter.New(orderClient, g.Order...),
		OrderItemQuery:               orderitemadapter.New(orderItemClient, g.OrderItem...),
		TransactionCommandRepository: NewTransactionCommandRepository(db),
		TransactionQueryRepository:   NewTransactionQueryRepository(db),
	}
}
