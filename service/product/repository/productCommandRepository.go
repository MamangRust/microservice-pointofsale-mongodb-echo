package repository

import (
	"context"
	"errors"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type productCommandRepository struct {
	db *mongo.Database
}

func NewProductCommandRepository(db *mongo.Database) *productCommandRepository {
	return &productCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *productCommandRepository) CreateProduct(ctx context.Context, request *requests.CreateProductRequest) (*models.Product, error) {
	productID, err := database.GetNextID(ctx, r.db, "products")
	if err != nil {
		return nil, product_errors.ErrCreateProduct
	}

	now := time.Now().UTC()
	weight := int32(request.Weight)
	product := &models.Product{
		ProductID:    productID,
		MerchantID:   int32(request.MerchantID),
		CategoryID:   int32(request.CategoryID),
		Name:         request.Name,
		Description:  &request.Description,
		Price:        int32(request.Price),
		CountInStock: int32(request.CountInStock),
		Brand:        &request.Brand,
		Weight:       &weight,
		SlugProduct:  request.SlugProduct,
		ImageProduct: &request.ImageProduct,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if _, err := r.db.Collection("products").InsertOne(ctx, product); err != nil {
		return nil, product_errors.ErrCreateProduct
	}
	return product, nil
}

func (r *productCommandRepository) UpdateProduct(ctx context.Context, request *requests.UpdateProductRequest) (*models.Product, error) {
	var product models.Product
	coll := r.db.Collection("products")
	err := coll.FindOne(ctx, bson.M{"_id": int32(*request.ProductID)}).Decode(&product)
	if err != nil {
		return nil, product_errors.ErrUpdateProduct
	}

	weight := int32(request.Weight)
	product.CategoryID = int32(request.CategoryID)
	product.Name = request.Name
	product.Description = &request.Description
	product.Price = int32(request.Price)
	product.CountInStock = int32(request.CountInStock)
	product.Brand = &request.Brand
	product.Weight = &weight
	product.ImageProduct = &request.ImageProduct
	product.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": product.ProductID}, product); err != nil {
		return nil, product_errors.ErrUpdateProduct
	}
	return &product, nil
}

func (r *productCommandRepository) UpdateProductCountStock(ctx context.Context, product_id int, stock int) (*models.Product, error) {
	var product models.Product
	coll := r.db.Collection("products")
	err := coll.FindOne(ctx, bson.M{"_id": int32(product_id)}).Decode(&product)
	if err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}

	product.CountInStock = int32(stock)
	product.UpdatedAt = timePtr(time.Now().UTC())

	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": product.ProductID}, product); err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	return &product, nil
}

func (r *productCommandRepository) DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return r.adjustProductCountStock(ctx, productID, -int32(quantity), true)
}

func (r *productCommandRepository) IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return r.adjustProductCountStock(ctx, productID, int32(quantity), false)
}

// adjustProductCountStock applies an atomic $inc to count_in_stock. When
// requireSufficient is true, the update only matches if the resulting stock
// stays non-negative so concurrent orders cannot oversell.
func (r *productCommandRepository) adjustProductCountStock(ctx context.Context, productID int, delta int32, requireSufficient bool) (*models.Product, error) {
	filter := bson.M{"_id": int32(productID)}
	if requireSufficient {
		filter["count_in_stock"] = bson.M{"$gte": -delta}
	}

	update := bson.M{
		"$inc": bson.M{"count_in_stock": delta},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var product models.Product
	err := r.db.Collection("products").FindOneAndUpdate(ctx, filter, update, opts).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) && requireSufficient {
			return nil, product_errors.ErrInsufficientProductStock
		}
		return nil, product_errors.ErrUpdateProductCountStock
	}
	return &product, nil
}

func (r *productCommandRepository) TrashedProduct(ctx context.Context, product_id int) (*models.Product, error) {
	var product models.Product
	coll := r.db.Collection("products")
	err := coll.FindOne(ctx, bson.M{"_id": int32(product_id)}).Decode(&product)
	if err != nil {
		return nil, product_errors.ErrTrashedProduct
	}

	now := time.Now().UTC()
	_, err = coll.UpdateOne(ctx, bson.M{"_id": product.ProductID}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	if err != nil {
		return nil, product_errors.ErrTrashedProduct
	}
	return &product, nil
}

func (r *productCommandRepository) RestoreProduct(ctx context.Context, product_id int) (*models.Product, error) {
	var product models.Product
	coll := r.db.Collection("products")
	err := coll.FindOne(ctx, bson.M{"_id": int32(product_id)}).Decode(&product)
	if err != nil {
		return nil, product_errors.ErrRestoreProduct
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": product.ProductID}, bson.M{
		"$unset": bson.M{"deleted_at": ""},
	})
	if err != nil {
		return nil, product_errors.ErrRestoreProduct
	}
	return &product, nil
}

func (r *productCommandRepository) DeleteProductPermanent(ctx context.Context, product_id int) (bool, error) {
	result, err := r.db.Collection("products").DeleteOne(ctx, bson.M{"_id": int32(product_id)})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}

func (r *productCommandRepository) RestoreAllProducts(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("products").UpdateMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	return true, err
}

func (r *productCommandRepository) DeleteAllProductPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.Collection("products").DeleteMany(ctx, bson.M{
		"deleted_at": bson.M{"$exists": true, "$ne": nil},
	})
	return true, err
}
