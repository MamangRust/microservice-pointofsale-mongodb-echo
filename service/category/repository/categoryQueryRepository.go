package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type categoryQueryRepository struct {
	db *mongo.Database
}

func NewCategoryQueryRepository(db *mongo.Database) CategoryQueryRepository {
	return &categoryQueryRepository{db: db}
}

func decodeCategoryFacetResult(cursor *mongo.Cursor, ctx context.Context) (totalCount int64, data []bson.M, err error) {
	var facetResults []struct {
		Metadata []struct {
			TotalCount int64 `bson:"total_count"`
		} `bson:"metadata"`
		Data []bson.M `bson:"data"`
	}
	if err = cursor.All(ctx, &facetResults); err != nil {
		return 0, nil, err
	}
	if len(facetResults) == 0 {
		return 0, nil, nil
	}
	if len(facetResults[0].Metadata) > 0 {
		totalCount = facetResults[0].Metadata[0].TotalCount
	}
	return totalCount, facetResults[0].Data, nil
}

func catBsonMToTimeStr(v bson.M, key string) *string {
	if val, ok := v[key]; ok {
		if dt, ok := val.(bson.DateTime); ok {
			s := dt.Time().Format(time.RFC3339)
			return &s
		}
	}
	return nil
}

func catBsonMToInt32(v bson.M, key string) int32 {
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

func bsonMToStrPtr(v bson.M, key string) *string {
	if val, ok := v[key]; ok {
		s := fmt.Sprintf("%v", val)
		return &s
	}
	return nil
}

func (r *categoryQueryRepository) FindAllCategory(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResult, *int, error) {
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

	cursor, err := r.db.Collection("categories").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, category_errors.ErrFindAllCategory
	}
	defer cursor.Close(ctx)

	totalCount, data, err := decodeCategoryFacetResult(cursor, ctx)
	if err != nil {
		return nil, nil, category_errors.ErrFindAllCategory
	}

	var results []*CategoryResult
	for _, c := range data {
		r := &CategoryResult{
			CategoryID:   catBsonMToInt32(c, "_id"),
			Name:         fmt.Sprintf("%v", c["name"]),
			Description:  bsonMToStrPtr(c, "description"),
			SlugCategory: bsonMToStrPtr(c, "slug_category"),
			TotalCount:   totalCount,
			CreatedAt:    catBsonMToTimeStr(c, "created_at"),
			UpdatedAt:    catBsonMToTimeStr(c, "updated_at"),
		}
		results = append(results, r)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *categoryQueryRepository) FindById(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection("categories").FindOne(ctx, bson.M{"_id": int32(category_id)}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrFindById
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByIdTrashed(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection("categories").FindOne(ctx, bson.M{"_id": int32(category_id)}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrFindById
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByNameAndId(ctx context.Context, req *requests.CategoryNameAndId) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection("categories").FindOne(ctx, bson.M{
		"name": req.Name,
		"_id":  bson.M{"$ne": int32(req.CategoryID)},
	}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrFindByNameAndId
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByName(ctx context.Context, name string) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection("categories").FindOne(ctx, bson.M{
		"name":       name,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&category)
	if err != nil {
		return nil, category_errors.ErrFindByName
	}
	return &category, nil
}

// FindByIds resolves many categories in a single round trip. Soft-deleted
// categories are skipped, matching FindByName/FindAllCategory.
func (r *categoryQueryRepository) FindByIds(ctx context.Context, category_ids []int) ([]*models.Category, error) {
	if len(category_ids) == 0 {
		return nil, nil
	}

	ids := make([]int32, 0, len(category_ids))
	for _, id := range category_ids {
		ids = append(ids, int32(id))
	}

	cursor, err := r.db.Collection("categories").Find(ctx, bson.M{
		"_id":        bson.M{"$in": ids},
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		return nil, category_errors.ErrFindByIds
	}
	defer cursor.Close(ctx)

	var categories []*models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, category_errors.ErrFindByIds
	}
	return categories, nil
}

func (r *categoryQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("categories").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("categories").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, category_errors.ErrFindByActive
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, nil, category_errors.ErrFindByActive
	}

	var results []*CategoryResultDeleteAt
	for _, c := range categories {
		cr := &CategoryResultDeleteAt{CategoryID: c.CategoryID, Name: c.Name, Description: c.Description, SlugCategory: c.SlugCategory, TotalCount: count}
		if c.CreatedAt != nil {
			s := c.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if c.UpdatedAt != nil {
			s := c.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		if c.DeletedAt != nil {
			s := c.DeletedAt.Format("2006-01-02 15:04:05")
			cr.DeletedAt = &s
		}
		results = append(results, cr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *categoryQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("categories").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("categories").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, category_errors.ErrFindByTrashed
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, nil, category_errors.ErrFindByTrashed
	}

	var results []*CategoryResultDeleteAt
	for _, c := range categories {
		cr := &CategoryResultDeleteAt{CategoryID: c.CategoryID, Name: c.Name, Description: c.Description, SlugCategory: c.SlugCategory, TotalCount: count}
		if c.CreatedAt != nil {
			s := c.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if c.UpdatedAt != nil {
			s := c.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		if c.DeletedAt != nil {
			s := c.DeletedAt.Format("2006-01-02 15:04:05")
			cr.DeletedAt = &s
		}
		results = append(results, cr)
	}
	total := int(count)
	return results, &total, nil
}
