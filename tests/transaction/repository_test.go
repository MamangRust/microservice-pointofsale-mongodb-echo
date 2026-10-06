package transaction_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	"github.com/stretchr/testify/suite"
)

type TransactionRepositoryTestSuite struct {
	tests.BaseTestSuite
	repo *repository.Repositories
}

func (s *TransactionRepositoryTestSuite) SetupSuite() {
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()

	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupOrderService()

	db := s.MongoDB()
	s.repo = repository.NewRepositories(db, nil, nil, nil, nil)
}

func (s *TransactionRepositoryTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *TransactionRepositoryTestSuite) TestTransactionLifecycle() {
	ctx := context.Background()

	userID := s.SeedUser(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, merchID, catID)
	orderID := s.SeedOrder(ctx, userID, merchID, prodID)

	pageReq := &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10}

	status := "paid"
	req := &requests.CreateTransactionRequest{
		OrderID:       orderID,
		CashierID:     userID,
		MerchantID:    merchID,
		PaymentMethod: "Credit Card",
		PaymentStatus: &status,
		Amount:        100000,
	}
	created, err := s.repo.TransactionCommandRepository.CreateTransaction(ctx, req)
	s.NoError(err)
	s.NotNil(created)
	s.Equal(int32(orderID), created.OrderID)
	txnID := int(created.TransactionID)

	found, err := s.repo.TransactionQueryRepository.FindById(ctx, txnID)
	s.NoError(err)
	s.Equal(created.Amount, found.Amount)

	foundByOrder, err := s.repo.TransactionQueryRepository.FindByOrderId(ctx, orderID)
	s.NoError(err)
	s.NotNil(foundByOrder)

	all, _, err := s.repo.TransactionQueryRepository.FindAllTransactions(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(all)

	active, _, err := s.repo.TransactionQueryRepository.FindByActive(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(active)

	updatedStatus := "settled"
	updateReq := &requests.UpdateTransactionRequest{
		TransactionID: &txnID,
		OrderID:       orderID,
		MerchantID:    merchID,
		PaymentMethod: "Debit Card",
		Amount:        120000,
		PaymentStatus: &updatedStatus,
	}
	updated, err := s.repo.TransactionCommandRepository.UpdateTransaction(ctx, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int32(120000), updated.Amount)

	trashed, err := s.repo.TransactionCommandRepository.TrashTransaction(ctx, txnID)
	s.NoError(err)
	s.NotNil(trashed)

	trashedItems, _, err := s.repo.TransactionQueryRepository.FindByTrashed(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(trashedItems)

	activeAfterTrash, _, err := s.repo.TransactionQueryRepository.FindByActive(ctx, pageReq)
	s.NoError(err)
	for _, item := range activeAfterTrash {
		s.NotEqual(txnID, int(item.TransactionID))
	}

	restored, err := s.repo.TransactionCommandRepository.RestoreTransaction(ctx, txnID)
	s.NoError(err)
	s.NotNil(restored)

	_, err = s.repo.TransactionCommandRepository.TrashTransaction(ctx, txnID)

	success, err := s.repo.TransactionCommandRepository.DeleteTransactionPermanently(ctx, txnID)
	s.NoError(err)
	s.True(success)

	status2 := "pending"
	second, _ := s.repo.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		CashierID: userID, OrderID: orderID, MerchantID: merchID,
		PaymentMethod: "Bank Transfer", PaymentStatus: &status2, Amount: 50000,
	})
	s.repo.TransactionCommandRepository.TrashTransaction(ctx, int(second.TransactionID))

	resRestore, err := s.repo.TransactionCommandRepository.RestoreAllTransactions(ctx)
	s.NoError(err)
	s.True(resRestore)

	s.repo.TransactionCommandRepository.TrashTransaction(ctx, int(second.TransactionID))
	resDelete, err := s.repo.TransactionCommandRepository.DeleteAllTransactionPermanent(ctx)
	s.NoError(err)
	s.True(resDelete)
}

func TestTransactionRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionRepositoryTestSuite))
}
