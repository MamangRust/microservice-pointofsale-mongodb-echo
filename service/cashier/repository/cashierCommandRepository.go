package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type cashierCommandRepository struct {
	db *mongo.Database
}

func NewCashierCommandRepository(db *mongo.Database) CashierCommandRepository {
	return &cashierCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *cashierCommandRepository) CreateCashier(ctx context.Context, request *requests.CreateCashierRequest) (*models.Cashier, error) {
	cashierID, err := database.GetNextID(ctx, r.db, "cashiers")
	if err != nil {
		return nil, cashier_errors.ErrCreateCashier
	}

	now := time.Now().UTC()
	cashier := &models.Cashier{
		CashierID:  cashierID,
		MerchantID: int32(request.MerchantID),
		UserID:     int32(request.UserID),
		Name:       request.Name,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	if _, err := r.db.Collection("cashiers").InsertOne(ctx, cashier); err != nil {
		return nil, cashier_errors.ErrCreateCashier
	}
	return cashier, nil
}

func (r *cashierCommandRepository) UpdateCashier(ctx context.Context, request *requests.UpdateCashierRequest) (*models.Cashier, error) {
	var cashier models.Cashier
	coll := r.db.Collection("cashiers")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.CashierID)}).Decode(&cashier)
	if err != nil {
		return nil, cashier_errors.ErrUpdateCashier
	}

	cashier.Name = request.Name
	cashier.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": cashier.CashierID}, cashier); err != nil {
		return nil, cashier_errors.ErrUpdateCashier
	}
	return &cashier, nil
}

func (r *cashierCommandRepository) TrashedCashier(ctx context.Context, cashier_id int) (*models.Cashier, error) {
	var cashier models.Cashier
	coll := r.db.Collection("cashiers")
	err := coll.FindOne(ctx, bson.M{"_id": int32(cashier_id)}).Decode(&cashier)
	if err != nil {
		return nil, cashier_errors.ErrTrashedCashier
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": cashier.CashierID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, cashier_errors.ErrTrashedCashier
	}
	return &cashier, nil
}

func (r *cashierCommandRepository) RestoreCashier(ctx context.Context, cashier_id int) (*models.Cashier, error) {
	var cashier models.Cashier
	coll := r.db.Collection("cashiers")
	err := coll.FindOne(ctx, bson.M{"_id": int32(cashier_id)}).Decode(&cashier)
	if err != nil {
		return nil, cashier_errors.ErrRestoreCashier
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": cashier.CashierID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, cashier_errors.ErrRestoreCashier
	}
	return &cashier, nil
}

func (r *cashierCommandRepository) DeleteCashierPermanent(ctx context.Context, cashier_id int) (bool, error) {
	_, err := r.db.Collection("cashiers").DeleteOne(ctx, bson.M{"_id": int32(cashier_id)})
	if err != nil {
		return false, cashier_errors.ErrDeleteCashierPermanent
	}
	return true, nil
}

func (r *cashierCommandRepository) RestoreAllCashier(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("cashiers").UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return false, cashier_errors.ErrRestoreAllCashiers
	}
	return true, nil
}

func (r *cashierCommandRepository) DeleteAllCashierPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("cashiers").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return false, cashier_errors.ErrDeleteAllCashiersPermanent
	}
	return true, nil
}
