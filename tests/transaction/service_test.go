package transaction_test

import (
	"context"
	"testing"

	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	trans_cache "github.com/MamangRust/microservice-point-of-sale-transacton/cache"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	"github.com/MamangRust/microservice-point-of-sale-transacton/service"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransactionServiceTestSuite struct {
	tests.BaseTestSuite
	svc *service.Service
}

func (s *TransactionServiceTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()

	s.SetupTransactionService()
	s.SetupCategoryService()

	db := s.MongoDB()

	mencache := trans_cache.NewMencache(s.GetCacheStore())
	repos := repository.NewRepositories(
		db,
		pbcashier.NewCashierServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]),
		pborder.NewOrderServiceClient(s.Conns["order"]),
		pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]),
	)

	s.svc = service.NewService(&service.Deps{
		Kafka:         nil,
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
}

func (s *TransactionServiceTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *TransactionServiceTestSuite) TestTransactionLifecycle() {
	ctx := context.Background()

	userID := s.SeedUser(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	categoryID := s.SeedCategory(ctx)
	productID := s.SeedProduct(ctx, merchantID, categoryID)
	orderID := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, orderID, productID)

	// Find cashier ID from MongoDB
	var cashierDoc bson.M
	err := s.Coll("cashiers").FindOne(ctx, bson.M{"user_id": int32(userID), "merchant_id": int32(merchantID)}).Decode(&cashierDoc)
	s.Require().NoError(err)
	cashierID := int(bsonMToInt32FromTx(cashierDoc, "_id"))

	req := &requests.CreateTransactionRequest{
		OrderID:       orderID,
		CashierID:     cashierID,
		MerchantID:    merchantID,
		PaymentMethod: "Transfer Bank",
		Amount:        1000000,
	}
	created, err := s.svc.TransactionCommand.CreateTransaction(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	transactionID := int(created.TransactionID)

	found, err := s.svc.TransactionQuery.FindById(ctx, transactionID)
	s.Require().NoError(err)
	s.Require().NotNil(found.PaymentStatus)
	s.Equal("success", *found.PaymentStatus)

	newPaymentMethod := "GOPAY"
	updateReq := &requests.UpdateTransactionRequest{
		TransactionID: &transactionID,
		OrderID:       orderID,
		CashierID:     cashierID,
		PaymentMethod: newPaymentMethod,
		Amount:        req.Amount,
	}
	updated, err := s.svc.TransactionCommand.UpdateTransaction(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(newPaymentMethod, updated.PaymentMethod)

	_, total, err := s.svc.TransactionQuery.FindAllTransactions(ctx, &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	_, err = s.svc.TransactionCommand.TrashedTransaction(ctx, transactionID)
	s.Require().NoError(err)

	_, totalTrashed, err := s.svc.TransactionQuery.FindByTrashed(ctx, &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	active, _, err := s.svc.TransactionQuery.FindByActive(ctx, &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, tx := range active {
		s.NotEqual(transactionID, int(tx.TransactionID))
	}

	_, err = s.svc.TransactionCommand.RestoreTransaction(ctx, transactionID)
	s.Require().NoError(err)

	_, err = s.svc.TransactionCommand.TrashedTransaction(ctx, transactionID)
	s.Require().NoError(err)
	success, err := s.svc.TransactionCommand.DeleteTransactionPermanently(ctx, transactionID)
	s.Require().NoError(err)
	s.True(success)

	o1 := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, o1, productID)

	o2 := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, o2, productID)

	s.svc.TransactionCommand.TrashedTransaction(ctx, o1)
	s.svc.TransactionCommand.TrashedTransaction(ctx, o2)

	resRestoreAll, err := s.svc.TransactionCommand.RestoreAllTransactions(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.TransactionCommand.TrashedTransaction(ctx, o1)
	s.svc.TransactionCommand.TrashedTransaction(ctx, o2)

	resDeleteAll, err := s.svc.TransactionCommand.DeleteAllTransactionPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func bsonMToInt32FromTx(d bson.M, key string) int32 {
	if val, ok := d[key]; ok {
		switch n := val.(type) {
		case int32:
			return n
		case int64:
			return int32(n)
		case float64:
			return int32(n)
		}
	}
	return 0
}

func TestTransactionServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionServiceTestSuite))
}
