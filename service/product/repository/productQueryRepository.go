package repository

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type productQueryRepository struct {
	db *mongo.Database
}

func bsonMToInt32Val(v bson.M, key string) int32 {
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

func NewProductQueryRepository(db *mongo.Database) *productQueryRepository {
	return &productQueryRepository{db: db}
}

func (r *productQueryRepository) FindAllProducts(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("products").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("products").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, product_errors.ErrFindAllProducts
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, nil, product_errors.ErrFindAllProducts
	}

	var results []*ProductResult
	for _, p := range products {
		pr := &ProductResult{
			ProductID:    p.ProductID,
			MerchantID:   p.MerchantID,
			CategoryID:   p.CategoryID,
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			CountInStock: p.CountInStock,
			Brand:        p.Brand,
			Weight:       p.Weight,
			SlugProduct:  p.SlugProduct,
			ImageProduct: p.ImageProduct,
			Barcode:      p.Barcode,
			TotalCount:   count,
		}
		if p.CreatedAt != nil {
			pr.CreatedAt = p.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt != nil {
			pr.UpdatedAt = p.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, pr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *productQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": false}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("products").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("products").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, product_errors.ErrFindByActive
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, nil, product_errors.ErrFindByActive
	}

	var results []*ProductResultDeleteAt
	for _, p := range products {
		pr := &ProductResultDeleteAt{
			ProductID: p.ProductID, MerchantID: p.MerchantID, CategoryID: p.CategoryID,
			Name: p.Name, Description: p.Description, Price: p.Price, CountInStock: p.CountInStock,
			Brand: p.Brand, Weight: p.Weight, SlugProduct: p.SlugProduct, ImageProduct: p.ImageProduct,
			Barcode: p.Barcode, TotalCount: count,
		}
		if p.CreatedAt != nil {
			pr.CreatedAt = p.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt != nil {
			pr.UpdatedAt = p.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		if p.DeletedAt != nil {
			pr.DeletedAt = p.DeletedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, pr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *productQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResultDeleteAt, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{"deleted_at": bson.M{"$exists": true, "$ne": nil}}
	if req.Search != "" {
		matchStage["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	count, _ := r.db.Collection("products").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("products").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, product_errors.ErrFindByTrashed
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, nil, product_errors.ErrFindByTrashed
	}

	var results []*ProductResultDeleteAt
	for _, p := range products {
		pr := &ProductResultDeleteAt{
			ProductID: p.ProductID, MerchantID: p.MerchantID, CategoryID: p.CategoryID,
			Name: p.Name, Description: p.Description, Price: p.Price, CountInStock: p.CountInStock,
			Brand: p.Brand, Weight: p.Weight, SlugProduct: p.SlugProduct, ImageProduct: p.ImageProduct,
			Barcode: p.Barcode, TotalCount: count,
		}
		if p.CreatedAt != nil {
			pr.CreatedAt = p.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt != nil {
			pr.UpdatedAt = p.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		if p.DeletedAt != nil {
			pr.DeletedAt = p.DeletedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, pr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *productQueryRepository) FindByMerchant(ctx context.Context, req *requests.ProductByMerchantRequest) ([]*ProductByMerchantResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{
		"merchant_id": int32(req.MerchantID),
		"deleted_at":  bson.M{"$exists": false},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$project", Value: bson.M{
			"_id": 1, "product_id": "$_id", "name": 1, "description": 1, "price": 1,
			"count_in_stock": 1, "brand": 1, "image_product": 1, "created_at": 1, "category_id": 1,
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": mongo.Pipeline{{{Key: "$count", Value: "total_count"}}},
			"data":     mongo.Pipeline{{{Key: "$skip", Value: offset}}, {{Key: "$limit", Value: limit}}},
		}}},
	}

	cursor, err := r.db.Collection("products").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, nil, product_errors.ErrFindByMerchant
	}
	defer cursor.Close(ctx)

	var facetResults []struct {
		Metadata []struct {
			TotalCount int64 `bson:"total_count"`
		} `bson:"metadata"`
		Data []bson.M `bson:"data"`
	}
	if err := cursor.All(ctx, &facetResults); err != nil {
		return nil, nil, product_errors.ErrFindByMerchant
	}

	var totalCount int64
	if len(facetResults) > 0 && len(facetResults[0].Metadata) > 0 {
		totalCount = facetResults[0].Metadata[0].TotalCount
	}

	var data []bson.M
	if len(facetResults) > 0 {
		data = facetResults[0].Data
	}

	var results []*ProductByMerchantResult
	for _, d := range data {
		createdAt := ""
		if v, ok := d["created_at"].(bson.DateTime); ok {
			createdAt = v.Time().Format("2006-01-02 15:04:05")
		}
		pr := &ProductByMerchantResult{
			TotalCount: totalCount, ProductID: bsonMToInt32Val(d, "_id"), Name: fmt.Sprintf("%v", d["name"]),
			Description: bsonMToStrPtr(d, "description"), Price: bsonMToInt32Val(d, "price"),
			CountInStock: bsonMToInt32Val(d, "count_in_stock"), Brand: bsonMToStrPtr(d, "brand"),
			ImageProduct: bsonMToStrPtr(d, "image_product"), CreatedAt: createdAt,
			CategoryID: bsonMToInt32Val(d, "category_id"),
		}
		results = append(results, pr)
	}
	total := int(totalCount)
	return results, &total, nil
}

// FindByCategoryID lists the products that belong to a category. The category
// name -> id resolution happens in the service layer through the Category
// adapter, so this repository stays Mongo-only.
func (r *productQueryRepository) FindByCategoryID(ctx context.Context, categoryID int32, req *requests.ProductByCategoryRequest) ([]*ProductByCategoryResult, *int, error) {
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	matchStage := bson.M{
		"category_id": categoryID,
		"deleted_at":  bson.M{"$exists": false},
	}

	count, _ := r.db.Collection("products").CountDocuments(ctx, matchStage)
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(offset).SetLimit(limit)
	cursor, err := r.db.Collection("products").Find(ctx, matchStage, opts)
	if err != nil {
		return nil, nil, product_errors.ErrFindByCategory
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, nil, product_errors.ErrFindByCategory
	}

	var results []*ProductByCategoryResult
	for _, p := range products {
		pr := &ProductByCategoryResult{
			TotalCount: count, ProductID: p.ProductID, MerchantID: p.MerchantID, CategoryID: p.CategoryID,
			SlugProduct: p.SlugProduct, Weight: p.Weight, Name: p.Name, Description: p.Description,
			Price: p.Price, CountInStock: p.CountInStock, Brand: p.Brand, ImageProduct: p.ImageProduct,
		}
		if p.CreatedAt != nil {
			pr.CreatedAt = p.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt != nil {
			pr.UpdatedAt = p.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		results = append(results, pr)
	}
	total := int(count)
	return results, &total, nil
}

func (r *productQueryRepository) FindById(ctx context.Context, product_id int) (*models.Product, error) {
	var product models.Product
	err := r.db.Collection("products").FindOne(ctx, bson.M{
		"_id":        int32(product_id),
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&product)
	if err != nil {
		return nil, product_errors.ErrFindById
	}
	return &product, nil
}

func (r *productQueryRepository) FindByIdTrashed(ctx context.Context, id int) (*models.Product, error) {
	var product models.Product
	err := r.db.Collection("products").FindOne(ctx, bson.M{
		"_id":        int32(id),
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}).Decode(&product)
	if err != nil {
		return nil, product_errors.ErrFindById
	}
	return &product, nil
}
