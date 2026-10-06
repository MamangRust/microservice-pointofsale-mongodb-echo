package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func timePtr(t time.Time) *time.Time { return &t }

type categoryCommandRepository struct {
	db *mongo.Database
}

func NewCategoryCommandRepository(db *mongo.Database) CategoryCommandRepository {
	return &categoryCommandRepository{db: db}
}

func (r *categoryCommandRepository) CreateCategory(ctx context.Context, request *requests.CreateCategoryRequest) (*models.Category, error) {
	catID, err := database.GetNextID(ctx, r.db, "categories")
	if err != nil {
		return nil, category_errors.ErrCreateCategory
	}

	now := time.Now().UTC()
	category := &models.Category{
		CategoryID:   catID,
		Name:         request.Name,
		Description:  &request.Description,
		SlugCategory: request.SlugCategory,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if _, err := r.db.Collection("categories").InsertOne(ctx, category); err != nil {
		return nil, category_errors.ErrCreateCategory
	}
	return category, nil
}

func (r *categoryCommandRepository) UpdateCategory(ctx context.Context, request *requests.UpdateCategoryRequest) (*models.Category, error) {
	var category models.Category
	coll := r.db.Collection("categories")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.CategoryID)}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrUpdateCategory
	}

	category.Name = request.Name
	category.Description = &request.Description
	category.SlugCategory = request.SlugCategory
	category.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": category.CategoryID}, category); err != nil {
		return nil, category_errors.ErrUpdateCategory
	}
	return &category, nil
}

func (r *categoryCommandRepository) TrashedCategory(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	coll := r.db.Collection("categories")
	err := coll.FindOne(ctx, bson.M{"_id": int32(category_id)}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrTrashedCategory
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": category.CategoryID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, category_errors.ErrTrashedCategory
	}
	return &category, nil
}

func (r *categoryCommandRepository) RestoreCategory(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	coll := r.db.Collection("categories")
	err := coll.FindOne(ctx, bson.M{"_id": int32(category_id)}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrRestoreCategory
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": category.CategoryID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, category_errors.ErrRestoreCategory
	}
	return &category, nil
}

func (r *categoryCommandRepository) DeleteCategoryPermanently(ctx context.Context, category_id int) (bool, error) {
	_, err := r.db.Collection("categories").DeleteOne(ctx, bson.M{"_id": int32(category_id)})
	if err != nil {
		return false, category_errors.ErrDeleteCategoryPermanently
	}
	return true, nil
}

func (r *categoryCommandRepository) RestoreAllCategories(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("categories").UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return false, category_errors.ErrRestoreAllCategories
	}
	return true, nil
}

func (r *categoryCommandRepository) DeleteAllPermanentCategories(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("categories").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	if err != nil {
		return false, category_errors.ErrDeleteAllPermanentCategories
	}
	return true, nil
}
