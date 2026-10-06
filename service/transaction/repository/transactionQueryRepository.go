package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type transactionQueryRepository struct {
	db *mongo.Database
}

func NewTransactionQueryRepository(db *mongo.Database) *transactionQueryRepository {
	return &transactionQueryRepository{db: db}
}

func fmtTxTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *transactionQueryRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["payment_method"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("transactions").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("transactions").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindAllTransactions
	}
	defer cursor.Close(ctx)

	var txs []models.Transaction
	if err := cursor.All(ctx, &txs); err != nil {
		return nil, nil, transaction_errors.ErrFindAllTransactions
	}

	var out []*TransactionResult
	for _, t := range txs {
		out = append(out, &TransactionResult{
			TransactionID: t.TransactionID, OrderID: t.OrderID, MerchantID: t.MerchantID,
			PaymentMethod: t.PaymentMethod, Amount: t.Amount, ChangeAmount: t.ChangeAmount,
			PaymentStatus: fmtTxStr(t.PaymentStatus), CreatedAt: fmtTxTime(t.CreatedAt), TotalCount: count,
		})
	}
	total := int(count)
	return out, &total, nil
}

func (r *transactionQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["payment_method"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("transactions").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("transactions").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindByActive
	}
	defer cursor.Close(ctx)

	var txs []models.Transaction
	if err := cursor.All(ctx, &txs); err != nil {
		return nil, nil, transaction_errors.ErrFindByActive
	}

	var out []*TransactionResultDeleteAt
	for _, t := range txs {
		out = append(out, &TransactionResultDeleteAt{
			TransactionID: t.TransactionID, OrderID: t.OrderID, MerchantID: t.MerchantID,
			PaymentMethod: t.PaymentMethod, Amount: t.Amount, ChangeAmount: t.ChangeAmount,
			PaymentStatus: fmtTxStr(t.PaymentStatus), CreatedAt: fmtTxTime(t.CreatedAt),
			DeletedAt: fmtTxTime(t.DeletedAt), TotalCount: count,
		})
	}
	total := int(count)
	return out, &total, nil
}

func (r *transactionQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["payment_method"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("transactions").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "deleted_at", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("transactions").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindByTrashed
	}
	defer cursor.Close(ctx)

	var txs []models.Transaction
	if err := cursor.All(ctx, &txs); err != nil {
		return nil, nil, transaction_errors.ErrFindByTrashed
	}

	var out []*TransactionResultDeleteAt
	for _, t := range txs {
		out = append(out, &TransactionResultDeleteAt{
			TransactionID: t.TransactionID, OrderID: t.OrderID, MerchantID: t.MerchantID,
			PaymentMethod: t.PaymentMethod, Amount: t.Amount, ChangeAmount: t.ChangeAmount,
			PaymentStatus: fmtTxStr(t.PaymentStatus), CreatedAt: fmtTxTime(t.CreatedAt),
			DeletedAt: fmtTxTime(t.DeletedAt), TotalCount: count,
		})
	}
	total := int(count)
	return out, &total, nil
}

func (r *transactionQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*TransactionByMerchantResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{
		"merchant_id": int32(req.MerchantID),
		"deleted_at":  bson.M{"$exists": false},
	}
	if req.Search != "" {
		matchStage["payment_method"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("transactions").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("transactions").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindByMerchant
	}
	defer cursor.Close(ctx)

	var txs []models.Transaction
	if err := cursor.All(ctx, &txs); err != nil {
		return nil, nil, transaction_errors.ErrFindByMerchant
	}

	var out []*TransactionByMerchantResult
	for _, t := range txs {
		out = append(out, &TransactionByMerchantResult{
			TransactionID: t.TransactionID, OrderID: t.OrderID, MerchantID: t.MerchantID,
			PaymentMethod: t.PaymentMethod, Amount: t.Amount, ChangeAmount: t.ChangeAmount,
			PaymentStatus: fmtTxStr(t.PaymentStatus), CreatedAt: fmtTxTime(t.CreatedAt), TotalCount: count,
		})
	}
	total := int(count)
	return out, &total, nil
}

func (r *transactionQueryRepository) FindById(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.Collection("transactions").FindOne(ctx, bson.M{
		"_id":         int32(transaction_id),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&t)
	if err != nil {
		return nil, transaction_errors.ErrFindById
	}
	return &t, nil
}

func (r *transactionQueryRepository) FindByOrderId(ctx context.Context, order_id int) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.Collection("transactions").FindOne(ctx, bson.M{
		"order_id":    int32(order_id),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&t)
	if err != nil {
		return nil, transaction_errors.ErrFindByOrderId
	}
	return &t, nil
}

func fmtTxStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseTxTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
	}
	return &t
}
