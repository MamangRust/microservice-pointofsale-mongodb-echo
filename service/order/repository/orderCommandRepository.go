package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type orderCommandRepository struct {
	db *mongo.Database
}

func NewOrderCommandRepository(db *mongo.Database) OrderCommandRepository {
	return &orderCommandRepository{db: db}
}

func ptrTime(t time.Time) *time.Time { return &t }

func (r *orderCommandRepository) DeleteOrder(ctx context.Context, orderID int) error {
	now := time.Now().UTC()
	_, err := r.db.Collection("orders").UpdateOne(ctx, bson.M{"_id": int32(orderID)}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return order_errors.ErrDeleteOrderPermanent
	}
	return nil
}

func (r *orderCommandRepository) CreateOrder(ctx context.Context, request *requests.CreateOrderRecordRequest) (*models.Order, error) {
	orderID, err := database.GetNextID(ctx, r.db, "orders")
	if err != nil {
		return nil, order_errors.ErrCreateOrder
	}

	now := time.Now().UTC()
	order := &models.Order{
		OrderID:    orderID,
		MerchantID: int32(request.MerchantID),
		CashierID:  int32(request.CashierID),
		TotalPrice: int64(request.TotalPrice),
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	if _, err := r.db.Collection("orders").InsertOne(ctx, order); err != nil {
		return nil, order_errors.ErrCreateOrder
	}
	return order, nil
}

func (r *orderCommandRepository) FindAllTrashed(ctx context.Context) ([]*models.Order, error) {
	cursor, err := r.db.Collection("orders").Find(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return nil, order_errors.ErrFindAllTrashed
	}
	defer cursor.Close(ctx)

	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, order_errors.ErrFindAllTrashed
	}
	return orders, nil
}

func (r *orderCommandRepository) UpdateOrder(ctx context.Context, request *requests.UpdateOrderRecordRequest) (*models.Order, error) {
	var order models.Order
	coll := r.db.Collection("orders")
	err := coll.FindOne(ctx, bson.M{"_id": int32(request.OrderID)}).Decode(&order)
	if err != nil {
		return nil, order_errors.ErrUpdateOrder
	}

	order.TotalPrice = int64(request.TotalPrice)
	order.UpdatedAt = ptrTime(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": order.OrderID}, order); err != nil {
		return nil, order_errors.ErrUpdateOrder
	}
	return &order, nil
}

func (r *orderCommandRepository) TrashedOrder(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	coll := r.db.Collection("orders")
	err := coll.FindOne(ctx, bson.M{"_id": int32(orderID)}).Decode(&order)
	if err != nil {
		return nil, order_errors.ErrTrashedOrder.WithInternal(err)
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": order.OrderID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, order_errors.ErrTrashedOrder.WithInternal(err)
	}
	return &order, nil
}

func (r *orderCommandRepository) RestoreOrder(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	coll := r.db.Collection("orders")
	err := coll.FindOne(ctx, bson.M{"_id": int32(orderID)}).Decode(&order)
	if err != nil {
		return nil, order_errors.ErrRestoreOrderNotFound
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": order.OrderID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, order_errors.ErrRestoreOrder.WithInternal(err)
	}
	return &order, nil
}

func (r *orderCommandRepository) DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error) {
	result, err := r.db.Collection("orders").DeleteOne(ctx, bson.M{"_id": int32(orderID)})
	if err != nil {
		return false, order_errors.ErrDeleteOrderPermanent.WithInternal(err)
	}
	if result.DeletedCount == 0 {
		return false, order_errors.ErrDeleteOrderPermanentNotFound
	}
	return true, nil
}

func (r *orderCommandRepository) DeleteAllOrderPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("orders").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return false, order_errors.ErrDeleteAllOrderPermanent.WithInternal(err)
	}
	return true, nil
}
