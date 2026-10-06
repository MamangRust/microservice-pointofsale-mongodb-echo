package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type orderItemCommandRepository struct {
	db *mongo.Database
}

func NewOrderItemCommandRepository(db *mongo.Database) OrderItemCommandRepository {
	return &orderItemCommandRepository{db: db}
}

func ptrTime(t time.Time) *time.Time { return &t }

func (r *orderItemCommandRepository) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	itemID, err := database.GetNextID(ctx, r.db, "order_items")
	if err != nil {
		return nil, orderitem_errors.ErrCreateOrderItem
	}

	now := time.Now().UTC()
	item := &models.OrderItem{
		OrderItemID: itemID,
		OrderID:     int32(req.OrderID),
		ProductID:   int32(req.ProductID),
		Quantity:    int32(req.Quantity),
		Price:       int64(req.Price),
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if _, err := r.db.Collection("order_items").InsertOne(ctx, item); err != nil {
		return nil, orderitem_errors.ErrCreateOrderItem
	}
	return item, nil
}

func (r *orderItemCommandRepository) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	var item models.OrderItem
	coll := r.db.Collection("order_items")
	if err := coll.FindOne(ctx, bson.M{"_id": int32(req.OrderItemID)}).Decode(&item); err != nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}

	item.ProductID = int32(req.ProductID)
	item.Quantity = int32(req.Quantity)
	item.Price = int64(req.Price)
	item.UpdatedAt = ptrTime(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": item.OrderItemID}, item); err != nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}
	return &item, nil
}

// DeleteOrderItem removes a single order item by its own ID.
//
// The Mongo _id IS the OrderItemID (see models.OrderItem), so the filter must
// use "_id". Filtering on "order_id" instead would match nothing — order item
// IDs and order IDs come from separate counters.
func (r *orderItemCommandRepository) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	if _, err := r.db.Collection("order_items").DeleteOne(ctx, bson.M{"_id": int32(orderItemID)}); err != nil {
		return orderitem_errors.ErrDeleteOrderItemPermanent
	}
	return nil
}
