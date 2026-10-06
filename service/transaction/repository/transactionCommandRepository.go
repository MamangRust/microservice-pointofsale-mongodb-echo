package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type transactionCommandRepository struct {
	db *mongo.Database
}

func NewTransactionCommandRepository(db *mongo.Database) *transactionCommandRepository {
	return &transactionCommandRepository{db: db}
}

func (r *transactionCommandRepository) CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*models.Transaction, error) {
	txID, err := database.GetNextID(ctx, r.db, "transactions")
	if err != nil {
		return nil, transaction_errors.ErrCreateTransaction
	}

	now := time.Now().UTC()
	t := &models.Transaction{
		TransactionID: txID,
		OrderID:       int32(request.OrderID),
		MerchantID:    int32(request.MerchantID),
		PaymentMethod: request.PaymentMethod,
		Amount:        int32(request.Amount),
		PaymentStatus: request.PaymentStatus,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	if _, err := r.db.Collection("transactions").InsertOne(ctx, t); err != nil {
		return nil, transaction_errors.ErrCreateTransaction
	}
	return t, nil
}

func (r *transactionCommandRepository) UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*models.Transaction, error) {
	var t models.Transaction
	coll := r.db.Collection("transactions")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.TransactionID)}).Decode(&t)
	if err != nil {
		return nil, transaction_errors.ErrUpdateTransaction
	}

	now := time.Now().UTC()
	t.MerchantID = int32(request.MerchantID)
	t.PaymentMethod = request.PaymentMethod
	t.Amount = int32(request.Amount)
	t.OrderID = int32(request.OrderID)
	t.PaymentStatus = request.PaymentStatus
	t.UpdatedAt = &now

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": t.TransactionID}, t); err != nil {
		return nil, transaction_errors.ErrUpdateTransaction
	}
	return &t, nil
}

func (r *transactionCommandRepository) TrashTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var t models.Transaction
	coll := r.db.Collection("transactions")
	err := coll.FindOne(ctx, bson.M{"_id": int32(transaction_id)}).Decode(&t)
	if err != nil {
		return nil, transaction_errors.ErrTrashTransaction
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": t.TransactionID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, transaction_errors.ErrTrashTransaction
	}
	t.DeletedAt = &now
	return &t, nil
}

func (r *transactionCommandRepository) RestoreTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var t models.Transaction
	coll := r.db.Collection("transactions")
	err := coll.FindOne(ctx, bson.M{"_id": int32(transaction_id)}).Decode(&t)
	if err != nil {
		return nil, transaction_errors.ErrRestoreTransaction
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": t.TransactionID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, transaction_errors.ErrRestoreTransaction
	}
	t.DeletedAt = nil
	return &t, nil
}

func (r *transactionCommandRepository) DeleteTransactionPermanently(ctx context.Context, transaction_id int) (bool, error) {
	result, err := r.db.Collection("transactions").DeleteOne(ctx, bson.M{"_id": int32(transaction_id)})
	if err != nil {
		return false, transaction_errors.ErrDeleteTransactionPermanently
	}
	if result.DeletedCount == 0 {
		return false, transaction_errors.ErrDeleteTransactionPermanently
	}
	return true, nil
}

func (r *transactionCommandRepository) RestoreAllTransactions(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("transactions").UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, transaction_errors.ErrRestoreAllTransactions
	}
	return true, nil
}

func (r *transactionCommandRepository) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("transactions").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return false, transaction_errors.ErrDeleteAllTransactionPermanent
	}
	return true, nil
}
