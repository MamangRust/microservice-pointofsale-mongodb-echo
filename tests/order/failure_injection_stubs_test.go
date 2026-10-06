package order_test

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// stubCashierRepo implements adapter/cashier.QueryRepository. In the
// failure-injection suite it is never reached (the merchant dependency fails
// first), so it returns a minimal valid row to keep the flow honest if the
// ordering ever changes.
type stubCashierRepo struct {
	cashierID int
}

func (s *stubCashierRepo) FindById(ctx context.Context, cashierID int) (*models.Cashier, error) {
	return &models.Cashier{CashierID: int32(cashierID), MerchantID: int32(s.cashierID), Name: "stub"}, nil
}

// stubProductRepo implements adapter/product.QueryRepository. Never reached
// in the failure-injection suite (merchant fails first).
type stubProductRepo struct{}

func (s *stubProductRepo) FindById(ctx context.Context, productID int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: 100}, nil
}

// stubProductCommandRepo implements adapter/product.CommandRepository. Stock
// mutations now go to the Product service over gRPC; in the failure-injection
// suite the merchant dependency fails first, so this stub is never reached.
type stubProductCommandRepo struct{}

func (s *stubProductCommandRepo) DecrementProductCountStock(_ context.Context, productID int, quantity int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: int32(100 - quantity)}, nil
}

func (s *stubProductCommandRepo) IncrementProductCountStock(_ context.Context, productID int, quantity int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: int32(100 + quantity)}, nil
}

// stubOrderItemRepo implements adapter/orderitem.QueryRepository and
// CommandRepository. Order-item writes now go to the OrderItem service over
// gRPC; in the failure-injection suite the merchant dependency fails first, so
// none of these methods is reached.
type stubOrderItemRepo struct {
	items []*models.OrderItem
}

func (s *stubOrderItemRepo) FindOrderItemByOrder(_ context.Context, orderID int) ([]*models.OrderItem, error) {
	return s.items, nil
}

func (s *stubOrderItemRepo) CalculateTotalPrice(_ context.Context, orderID int) (*int32, error) {
	var total int32
	for _, item := range s.items {
		total += item.Quantity * int32(item.Price)
	}
	return &total, nil
}

func (s *stubOrderItemRepo) CreateOrderItem(_ context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	return &models.OrderItem{
		OrderItemID: int32(len(s.items) + 1),
		OrderID:     int32(req.OrderID),
		ProductID:   int32(req.ProductID),
		Quantity:    int32(req.Quantity),
		Price:       int64(req.Price),
	}, nil
}

func (s *stubOrderItemRepo) UpdateOrderItem(_ context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	return &models.OrderItem{
		OrderItemID: int32(req.OrderItemID),
		OrderID:     int32(req.OrderID),
		ProductID:   int32(req.ProductID),
		Quantity:    int32(req.Quantity),
		Price:       int64(req.Price),
	}, nil
}

func (s *stubOrderItemRepo) DeleteOrderItem(_ context.Context, orderItemID int) error {
	return nil
}
