package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type orderItemQueryRepository struct {
	db *mongo.Database
}

func NewOrderItemQueryRepository(db *mongo.Database) OrderItemQueryRepository {
	return &orderItemQueryRepository{db: db}
}

// FindAllOrderItems returns one page of active order items.
//
// Order item IDs live in `_id` (see models.OrderItem), so the sort key is `_id`.
// Sorting on an `order_item_id` field — which is never written — left the order
// undefined and made skip/limit pagination unreliable, which matters for the
// stats backfill that walks every page.
func (r *orderItemQueryRepository) FindAllOrderItems(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}

	count, _ := r.db.Collection("order_items").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("order_items").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}
	defer cursor.Close(ctx)

	var items []models.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}

	var results []*OrderItemResult
	for _, item := range items {
		cr := &OrderItemResult{
			OrderItemID: item.OrderItemID, OrderID: item.OrderID, ProductID: item.ProductID,
			Quantity: item.Quantity, Price: int32(item.Price), TotalCount: count,
		}
		if item.CreatedAt != nil {
			s := item.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if item.UpdatedAt != nil {
			s := item.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		results = append(results, cr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderItemQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}

	count, _ := r.db.Collection("order_items").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("order_items").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}
	defer cursor.Close(ctx)

	var items []models.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}

	var results []*OrderItemResultDeleteAt
	for _, item := range items {
		cr := &OrderItemResultDeleteAt{
			OrderItemID: item.OrderItemID, OrderID: item.OrderID, ProductID: item.ProductID,
			Quantity: item.Quantity, Price: int32(item.Price), TotalCount: count,
		}
		if item.CreatedAt != nil {
			s := item.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if item.UpdatedAt != nil {
			s := item.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		if item.DeletedAt != nil {
			s := item.DeletedAt.Format("2006-01-02 15:04:05")
			cr.DeletedAt = &s
		}
		results = append(results, cr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderItemQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}

	count, _ := r.db.Collection("order_items").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("order_items").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}
	defer cursor.Close(ctx)

	var items []models.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}

	var results []*OrderItemResultDeleteAt
	for _, item := range items {
		cr := &OrderItemResultDeleteAt{
			OrderItemID: item.OrderItemID, OrderID: item.OrderID, ProductID: item.ProductID,
			Quantity: item.Quantity, Price: int32(item.Price), TotalCount: count,
		}
		if item.CreatedAt != nil {
			s := item.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if item.UpdatedAt != nil {
			s := item.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		if item.DeletedAt != nil {
			s := item.DeletedAt.Format("2006-01-02 15:04:05")
			cr.DeletedAt = &s
		}
		results = append(results, cr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderItemQueryRepository) FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	cursor, err := r.db.Collection("order_items").Find(ctx, bson.M{
		"order_id": int32(orderID),
	})
	if err != nil {
		return nil, orderitem_errors.ErrFindOrderItemByOrder
	}
	defer cursor.Close(ctx)

	var items []models.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, orderitem_errors.ErrFindOrderItemByOrder
	}

	var result []*models.OrderItem
	for i := range items {
		result = append(result, &items[i])
	}
	return result, nil
}
