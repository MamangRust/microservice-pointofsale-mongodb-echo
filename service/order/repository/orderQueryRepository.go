package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type orderQueryRepository struct {
	db *mongo.Database
}

func NewOrderQueryRepository(db *mongo.Database) OrderQueryRepository {
	return &orderQueryRepository{db: db}
}

func (r *orderQueryRepository) findOrders(ctx context.Context, matchStage bson.M, req interface{}) ([]*OrderResult, *int, error) {
	var offset, limit int64
	var search string
	switch v := req.(type) {
	case *requests.FindAllOrders:
		offset = int64((v.Page - 1) * v.PageSize)
		limit = int64(v.PageSize)
		search = v.Search
	case *requests.FindAllOrderMerchant:
		offset = int64((v.Page - 1) * v.PageSize)
		limit = int64(v.PageSize)
		search = v.Search
	}

	if search != "" {
		matchStage["$or"] = []bson.M{
			{"_id": bson.M{"$gte": 0, "$lte": 9999999}},
		}
	}

	count, _ := r.db.Collection("orders").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("orders").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, order_errors.ErrFindAllOrders
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, nil, order_errors.ErrFindAllOrders
	}

	var results []*OrderResult
	for _, o := range orders {
		or := &OrderResult{
			OrderID: o.OrderID, MerchantID: o.MerchantID, CashierID: o.CashierID, TotalPrice: o.TotalPrice, TotalCount: count,
		}
		if o.CreatedAt != nil {
			or.CreatedAt = o.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if o.UpdatedAt != nil {
			or.UpdatedAt = o.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, or)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderQueryRepository) FindAllOrders(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResult, *int, error) {
	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	return r.findOrders(ctx, matchStage, req)
}

func (r *orderQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}

	count, _ := r.db.Collection("orders").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("orders").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, order_errors.ErrFindByActive
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, nil, order_errors.ErrFindByActive
	}

	var results []*OrderResultDeleteAt
	for _, o := range orders {
		or := &OrderResultDeleteAt{
			OrderID: o.OrderID, MerchantID: o.MerchantID, CashierID: o.CashierID, TotalPrice: o.TotalPrice, TotalCount: count,
		}
		if o.CreatedAt != nil {
			or.CreatedAt = o.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if o.UpdatedAt != nil {
			or.UpdatedAt = o.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		if o.DeletedAt != nil {
			or.DeletedAt = o.DeletedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, or)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}

	count, _ := r.db.Collection("orders").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("orders").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, order_errors.ErrFindByTrashed
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, nil, order_errors.ErrFindByTrashed
	}

	var results []*OrderResultDeleteAt
	for _, o := range orders {
		or := &OrderResultDeleteAt{
			OrderID: o.OrderID, MerchantID: o.MerchantID, CashierID: o.CashierID, TotalPrice: o.TotalPrice, TotalCount: count,
		}
		if o.CreatedAt != nil {
			or.CreatedAt = o.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if o.UpdatedAt != nil {
			or.UpdatedAt = o.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		if o.DeletedAt != nil {
			or.DeletedAt = o.DeletedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, or)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*OrderResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{
		"merchant_id": int32(req.MerchantID),
		"deleted_at":  bson.M{"$exists": false},
	}

	count, _ := r.db.Collection("orders").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("orders").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, order_errors.ErrFindByMerchant
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, nil, order_errors.ErrFindByMerchant
	}

	var results []*OrderResult
	for _, o := range orders {
		or := &OrderResult{
			OrderID: o.OrderID, MerchantID: o.MerchantID, CashierID: o.CashierID, TotalPrice: o.TotalPrice, TotalCount: count,
		}
		if o.CreatedAt != nil {
			or.CreatedAt = o.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if o.UpdatedAt != nil {
			or.UpdatedAt = o.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, or)
	}
	total := int(count)
	return results, &total, nil
}

func (r *orderQueryRepository) FindById(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	err := r.db.Collection("orders").FindOne(ctx, bson.M{
		"_id":         int32(orderID),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&order)
	if err != nil {
		return nil, order_errors.ErrFindById
	}
	return &order, nil
}

func (r *orderQueryRepository) FindByTrashedId(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	err := r.db.Collection("orders").FindOne(ctx, bson.M{
		"_id":         int32(orderID),
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}).Decode(&order)
	if err != nil {
		return nil, order_errors.ErrFindByTrashedId
	}
	return &order, nil
}
