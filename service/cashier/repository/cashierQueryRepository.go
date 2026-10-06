package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type cashierQueryRepository struct {
	db *mongo.Database
}

func NewCashierQueryRepository(db *mongo.Database) CashierQueryRepository {
	return &cashierQueryRepository{db: db}
}

func decodeCashierFacetResult(cursor *mongo.Cursor, ctx context.Context) (totalCount int64, data []bson.M, err error) {
	var results []struct {
		Metadata []struct {
			TotalCount int64 `bson:"total_count"`
		} `bson:"metadata"`
		Data []bson.M `bson:"data"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, nil, err
	}
	if len(results) == 0 {
		return 0, nil, nil
	}
	if len(results[0].Metadata) > 0 {
		totalCount = results[0].Metadata[0].TotalCount
	}
	return totalCount, results[0].Data, nil
}

func cashierBsonMToTimeStr(v bson.M, key string) *string {
	if val, ok := v[key]; ok {
		if dt, ok := val.(bson.DateTime); ok {
			s := dt.Time().Format(time.RFC3339)
			return &s
		}
	}
	return nil
}

func cashierBsonMToInt32(v bson.M, key string) int32 {
	if val, ok := v[key]; ok {
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

func (r *cashierQueryRepository) FindAllCashiers(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{{{Key: "$count", Value: "total_count"}}},
			"data":     mongo.Pipeline{{{Key: "$skip", Value: offset}}, {{Key: "$limit", Value: limit}}},
		}}},
	}

	cursor, err := r.db.Collection("cashiers").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindAllCashiers
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeCashierFacetResult(cursor, ctx)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindAllCashiers
	}

	var results []*CashierResult
	for _, d := range data {
		cr := &CashierResult{
			CashierID:  cashierBsonMToInt32(d, "_id"),
			MerchantID: cashierBsonMToInt32(d, "merchant_id"),
			UserID:     cashierBsonMToInt32(d, "user_id"),
			Name:       fmt.Sprintf("%v", d["name"]),
			TotalCount: totalCount,
			CreatedAt:  cashierBsonMToTimeStr(d, "created_at"),
			UpdatedAt:  cashierBsonMToTimeStr(d, "updated_at"),
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *cashierQueryRepository) FindById(ctx context.Context, cashier_id int) (*models.Cashier, error) {
	var cashier models.Cashier
	err := r.db.Collection("cashiers").FindOne(ctx, bson.M{"_id": int32(cashier_id)}).Decode(&cashier)
	if err != nil {
		return nil, cashier_errors.ErrFindCashierById
	}
	return &cashier, nil
}

func (r *cashierQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{{{Key: "$count", Value: "total_count"}}},
			"data":     mongo.Pipeline{{{Key: "$skip", Value: offset}}, {{Key: "$limit", Value: limit}}},
		}}},
	}

	cursor, err := r.db.Collection("cashiers").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindActiveCashiers
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeCashierFacetResult(cursor, ctx)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindActiveCashiers
	}

	var results []*CashierResultDeleteAt
	for _, d := range data {
		cr := &CashierResultDeleteAt{
			CashierID:  cashierBsonMToInt32(d, "_id"),
			MerchantID: cashierBsonMToInt32(d, "merchant_id"),
			UserID:     cashierBsonMToInt32(d, "user_id"),
			Name:       fmt.Sprintf("%v", d["name"]),
			TotalCount: totalCount,
			CreatedAt:  cashierBsonMToTimeStr(d, "created_at"),
			UpdatedAt:  cashierBsonMToTimeStr(d, "updated_at"),
			DeletedAt:  cashierBsonMToTimeStr(d, "deleted_at"),
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *cashierQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{{{Key: "$count", Value: "total_count"}}},
			"data":     mongo.Pipeline{{{Key: "$skip", Value: offset}}, {{Key: "$limit", Value: limit}}},
		}}},
	}

	cursor, err := r.db.Collection("cashiers").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindTrashedCashiers
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeCashierFacetResult(cursor, ctx)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindTrashedCashiers
	}

	var results []*CashierResultDeleteAt
	for _, d := range data {
		cr := &CashierResultDeleteAt{
			CashierID:  cashierBsonMToInt32(d, "_id"),
			MerchantID: cashierBsonMToInt32(d, "merchant_id"),
			UserID:     cashierBsonMToInt32(d, "user_id"),
			Name:       fmt.Sprintf("%v", d["name"]),
			TotalCount: totalCount,
			CreatedAt:  cashierBsonMToTimeStr(d, "created_at"),
			UpdatedAt:  cashierBsonMToTimeStr(d, "updated_at"),
			DeletedAt:  cashierBsonMToTimeStr(d, "deleted_at"),
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *cashierQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*CashierResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{
		"merchant_id": int32(req.MerchantID),
		"deleted_at":  bson.M{"$exists": false},
	}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, err := r.db.Collection("cashiers").CountDocuments(ctx, matchStage)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindCashiersByMerchant
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("cashiers").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, cashier_errors.ErrFindCashiersByMerchant
	}
	defer cursor.Close(ctx)

	var cashiers []models.Cashier
	if err := cursor.All(ctx, &cashiers); err != nil {
		return nil, nil, cashier_errors.ErrFindCashiersByMerchant
	}

	var results []*CashierResult
	for _, c := range cashiers {
		cr := &CashierResult{CashierID: c.CashierID, MerchantID: c.MerchantID, UserID: c.UserID, Name: c.Name, TotalCount: count}
		if c.CreatedAt != nil {
			s := c.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if c.UpdatedAt != nil {
			s := c.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		results = append(results, cr)
	}
	total := int(count)
	return results, &total, nil
}
