package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

// ── Category ──────────────────────────────────────────────────────────────

type categorySeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewCategorySeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *categorySeeder {
	return &categorySeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *categorySeeder) Seed() error {
	coll := r.dbs.Catalog.Collection("categories")

	categoryNames := []string{
		"Electronics", "Clothing", "Groceries", "Toys", "Home & Kitchen",
		"Books", "Beauty & Health", "Sports & Outdoors", "Automotive", "Furniture",
	}
	categoryDescriptions := []string{
		"Best electronics products", "Latest fashion trends", "Fresh groceries",
		"Fun toys for kids", "Essentials for home & kitchen",
		"Books for all ages", "Beauty and health products",
		"Outdoor sports equipment", "Automotive accessories", "Stylish furniture",
	}

	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		name := categoryNames[i%len(categoryNames)]
		description := categoryDescriptions[i%len(categoryDescriptions)]
		slugCategory := fmt.Sprintf("%s-%d", name, i+1)

		catID, err := database.GetNextID(r.ctx, r.dbs.Catalog, "categories")
		if err != nil {
			return err
		}

		doc := bson.M{
			"_id":           catID,
			"name":          name,
			"description":   description,
			"slug_category": slugCategory,
			"created_at":    now,
			"updated_at":    now,
		}

		if _, err := coll.InsertOne(r.ctx, doc); err != nil {
			r.logger.Error("Failed to create category:", zap.Error(err))
			return err
		}
		r.logger.Debug("Category created:", zap.String("slug", slugCategory))
	}

	r.logger.Info("Category seeding completed successfully.", zap.Int("count", 10))
	return nil
}

// ── Cashier ───────────────────────────────────────────────────────────────

type cashierSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewCashierSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *cashierSeeder {
	return &cashierSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *cashierSeeder) Seed() error {
	merchantsColl := r.dbs.Merchant.Collection("merchants")
	usersColl := r.dbs.Identity.Collection("users")
	cashiersColl := r.dbs.Merchant.Collection("cashiers")

	merchantCursor, err := merchantsColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer merchantCursor.Close(r.ctx)
	var merchants []bson.M
	if err := merchantCursor.All(r.ctx, &merchants); err != nil {
		return err
	}

	userCursor, err := usersColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer userCursor.Close(r.ctx)
	var users []bson.M
	if err := userCursor.All(r.ctx, &users); err != nil {
		return err
	}

	if len(merchants) == 0 || len(users) == 0 {
		return fmt.Errorf("no merchants or users found")
	}

	now := time.Now().UTC()
	for i := 1; i <= 10; i++ {
		merchant := merchants[i%len(merchants)]
		user := users[i%len(users)]

		cashierID, err := database.GetNextID(r.ctx, r.dbs.Merchant, "cashiers")
		if err != nil {
			return err
		}

		doc := bson.M{
			"_id":         cashierID,
			"merchant_id": merchant["_id"],
			"user_id":     user["_id"],
			"name":        fmt.Sprintf("Cashier %d", i),
			"created_at":  now,
			"updated_at":  now,
		}

		if _, err := cashiersColl.InsertOne(r.ctx, doc); err != nil {
			return err
		}
	}

	r.logger.Info("Cashier seeding completed successfully.", zap.Int("count", 10))
	return nil
}

// ── Merchant ──────────────────────────────────────────────────────────────

type merchantSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewMerchantSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *merchantSeeder {
	return &merchantSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *merchantSeeder) Seed() error {
	usersColl := r.dbs.Identity.Collection("users")
	merchantsColl := r.dbs.Merchant.Collection("merchants")

	userCursor, err := usersColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer userCursor.Close(r.ctx)
	var users []bson.M
	if err := userCursor.All(r.ctx, &users); err != nil {
		return err
	}

	if len(users) == 0 {
		return fmt.Errorf("no users found")
	}

	now := time.Now().UTC()
	for i := 1; i <= 10; i++ {
		userID := users[i%len(users)]["_id"].(int32)

		merchantID, err := database.GetNextID(r.ctx, r.dbs.Merchant, "merchants")
		if err != nil {
			return err
		}

		doc := bson.M{
			"_id":           merchantID,
			"user_id":       userID,
			"name":          fmt.Sprintf("Toko %d", i),
			"description":   fmt.Sprintf("Deskripsi untuk Toko %d", i),
			"address":       fmt.Sprintf("Jl. Toko %d", i),
			"contact_email": fmt.Sprintf("toko%d@example.com", i),
			"contact_phone": fmt.Sprintf("0812345678%d", i),
			"status":        "active",
			"created_at":    now,
			"updated_at":    now,
		}

		if _, err := merchantsColl.InsertOne(r.ctx, doc); err != nil {
			return err
		}
	}

	r.logger.Info("Merchant seeding completed successfully.", zap.Int("count", 10))
	return nil
}

// ── Product ───────────────────────────────────────────────────────────────

type productSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewProductSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *productSeeder {
	return &productSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *productSeeder) Seed() error {
	merchantsColl := r.dbs.Merchant.Collection("merchants")
	categoriesColl := r.dbs.Catalog.Collection("categories")
	productsColl := r.dbs.Catalog.Collection("products")

	merchantCursor, err := merchantsColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer merchantCursor.Close(r.ctx)
	var merchants []bson.M
	if err := merchantCursor.All(r.ctx, &merchants); err != nil {
		return err
	}

	catCursor, err := categoriesColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer catCursor.Close(r.ctx)
	var categories []bson.M
	if err := catCursor.All(r.ctx, &categories); err != nil {
		return err
	}

	if len(merchants) == 0 || len(categories) == 0 {
		return fmt.Errorf("no merchants or categories found")
	}

	productNames := []string{
		"Smartphone", "Laptop", "Wireless Earbuds", "Gaming Mouse", "Mechanical Keyboard",
		"Smartwatch", "Power Bank", "Bluetooth Speaker", "External Hard Drive", "Monitor",
	}
	brands := []string{"Samsung", "Apple", "Sony", "Logitech", "Razer", "Xiaomi", "HP", "Dell", "Acer", "Asus"}
	images := []string{
		"image1.jpg", "image2.jpg", "image3.jpg", "image4.jpg", "image5.jpg",
		"image6.jpg", "image7.jpg", "image8.jpg", "image9.jpg", "image10.jpg",
	}

	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		merchant := merchants[i%len(merchants)]
		category := categories[i%len(categories)]
		name := productNames[i%len(productNames)]
		brand := brands[i%len(brands)]
		price := int32(50000 + i*100000)
		countInStock := int32(50 + i*10)
		weight := int32(200 + i*500)
		slug := fmt.Sprintf("%s-%d", name, i+1)
		image := images[i%len(images)]
		barcode := fmt.Sprintf("BC-%07d", i+1)

		productID, err := database.GetNextID(r.ctx, r.dbs.Catalog, "products")
		if err != nil {
			return err
		}

		doc := bson.M{
			"_id":            productID,
			"merchant_id":    merchant["_id"],
			"category_id":    category["_id"],
			"name":           name,
			"description":    fmt.Sprintf("Description for %s", name),
			"price":          price,
			"count_in_stock": countInStock,
			"brand":          brand,
			"weight":         weight,
			"slug_product":   slug,
			"image_product":  image,
			"barcode":        barcode,
			"created_at":     now,
			"updated_at":     now,
		}

		if _, err := productsColl.InsertOne(r.ctx, doc); err != nil {
			return err
		}
	}

	r.logger.Info("Product seeding completed successfully.", zap.Int("count", 10))
	return nil
}

// ── Order ─────────────────────────────────────────────────────────────────

type orderSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewOrderSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *orderSeeder {
	return &orderSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *orderSeeder) Seed() error {
	merchantsColl := r.dbs.Merchant.Collection("merchants")
	cashiersColl := r.dbs.Merchant.Collection("cashiers")
	productsColl := r.dbs.Catalog.Collection("products")
	ordersColl := r.dbs.Sales.Collection("orders")
	orderItemsColl := r.dbs.Sales.Collection("order_items")

	merchantCursor, err := merchantsColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer merchantCursor.Close(r.ctx)
	var merchants []bson.M
	if err := merchantCursor.All(r.ctx, &merchants); err != nil {
		return err
	}

	cashierCursor, err := cashiersColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer cashierCursor.Close(r.ctx)
	var cashiers []bson.M
	if err := cashierCursor.All(r.ctx, &cashiers); err != nil {
		return err
	}

	if len(merchants) == 0 || len(cashiers) == 0 {
		return fmt.Errorf("no merchants or cashiers found")
	}

	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		merchant := merchants[i%len(merchants)]
		cashier := cashiers[i%len(cashiers)]
		merchantID := merchant["_id"].(int32)
		totalPrice := int64(50000 + i*100000)

		orderID, err := database.GetNextID(r.ctx, r.dbs.Sales, "orders")
		if err != nil {
			return err
		}

		orderDoc := bson.M{
			"_id":         orderID,
			"merchant_id": merchantID,
			"cashier_id":  cashier["_id"],
			"total_price": totalPrice,
			"created_at":  now,
			"updated_at":  now,
		}

		if _, err := ordersColl.InsertOne(r.ctx, orderDoc); err != nil {
			return err
		}

		// Fetch products for this merchant
		prodCursor, err := productsColl.Find(r.ctx, bson.M{
			"merchant_id": merchantID,
			"deleted_at":  bson.M{"$exists": false},
		})
		if err != nil {
			return err
		}
		var products []bson.M
		if err := prodCursor.All(r.ctx, &products); err != nil {
			return err
		}
		prodCursor.Close(r.ctx)

		if len(products) == 0 {
			continue
		}

		for j := 0; j < 3; j++ {
			product := products[j%len(products)]
			quantity := int32(j + 1)
			price := int64(product["price"].(int32)) * int64(quantity)

			orderItemID, err := database.GetNextID(r.ctx, r.dbs.Sales, "order_items")
			if err != nil {
				return err
			}

			itemDoc := bson.M{
				"_id":        orderItemID,
				"order_id":   orderID,
				"product_id": product["_id"],
				"quantity":   quantity,
				"price":      price,
				"created_at": now,
				"updated_at": now,
			}

			if _, err := orderItemsColl.InsertOne(r.ctx, itemDoc); err != nil {
				return err
			}
		}
	}

	r.logger.Info("Order seeding completed successfully.", zap.Int("count", 10))
	return nil
}

// ── Transaction ───────────────────────────────────────────────────────────

type transactionSeeder struct {
	dbs    Databases
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewTransactionSeeder(dbs Databases, ctx context.Context, logger logger.LoggerInterface) *transactionSeeder {
	return &transactionSeeder{dbs: dbs, ctx: ctx, logger: logger}
}

func (r *transactionSeeder) Seed() error {
	ordersColl := r.dbs.Sales.Collection("orders")
	merchantsColl := r.dbs.Merchant.Collection("merchants")
	transactionsColl := r.dbs.Sales.Collection("transactions")

	orderCursor, err := ordersColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer orderCursor.Close(r.ctx)
	var orders []bson.M
	if err := orderCursor.All(r.ctx, &orders); err != nil {
		return err
	}

	merchantCursor, err := merchantsColl.Find(r.ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer merchantCursor.Close(r.ctx)
	var merchants []bson.M
	if err := merchantCursor.All(r.ctx, &merchants); err != nil {
		return err
	}

	if len(orders) == 0 || len(merchants) == 0 {
		return fmt.Errorf("no orders or merchants found")
	}

	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		selectedMerchant := merchants[i%len(merchants)]
		selectedOrder := orders[i%len(orders)]
		amount := int32(100 + i)
		changeAmount := int32(5 + i)

		txID, err := database.GetNextID(r.ctx, r.dbs.Sales, "transactions")
		if err != nil {
			return err
		}

		doc := bson.M{
			"_id":            txID,
			"order_id":       selectedOrder["_id"],
			"merchant_id":    selectedMerchant["_id"],
			"payment_method": "Credit Card",
			"amount":         amount,
			"change_amount":  &changeAmount,
			"payment_status": "Completed",
			"created_at":     now,
			"updated_at":     now,
		}

		if _, err := transactionsColl.InsertOne(r.ctx, doc); err != nil {
			return err
		}
	}

	r.logger.Info("Successfully seeded 10 transactions.", zap.Int("count", 10))
	return nil
}
