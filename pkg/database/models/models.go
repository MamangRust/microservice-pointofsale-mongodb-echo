package models

import "time"

type Role struct {
	RoleID    int32      `bson:"_id" json:"role_id"`
	RoleName  string     `bson:"role_name" json:"role_name"`
	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type User struct {
	UserID           int32      `bson:"_id" json:"user_id"`
	Firstname        string     `bson:"firstname" json:"firstname"`
	Lastname         string     `bson:"lastname" json:"lastname"`
	Email            string     `bson:"email" json:"email"`
	Password         string     `bson:"password" json:"password"`
	VerificationCode string     `bson:"verification_code" json:"verification_code"`
	IsVerified       *bool      `bson:"is_verified" json:"is_verified"`
	CreatedAt        *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt        *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt        *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type UserRole struct {
	UserRoleID int32      `bson:"_id" json:"user_role_id"`
	UserID     int32      `bson:"user_id" json:"user_id"`
	RoleID     int32      `bson:"role_id" json:"role_id"`
	CreatedAt  *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type RefreshToken struct {
	RefreshTokenID int32      `bson:"_id" json:"refresh_token_id"`
	UserID         int32      `bson:"user_id" json:"user_id"`
	Token          string     `bson:"token" json:"token"`
	Expiration     time.Time  `bson:"expiration" json:"expiration"`
	CreatedAt      *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type ResetToken struct {
	ResetTokenID int32     `bson:"_id" json:"id"`
	UserID       int64     `bson:"user_id" json:"user_id"`
	Token        string    `bson:"token" json:"token"`
	ExpiryDate   time.Time `bson:"expiry_date" json:"expiry_date"`
}

type Cashier struct {
	CashierID  int32      `bson:"_id" json:"cashier_id"`
	MerchantID int32      `bson:"merchant_id" json:"merchant_id"`
	UserID     int32      `bson:"user_id" json:"user_id"`
	Name       string     `bson:"name" json:"name"`
	CreatedAt  *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type Merchant struct {
	MerchantID   int32      `bson:"_id" json:"merchant_id"`
	UserID       int32      `bson:"user_id" json:"user_id"`
	Name         string     `bson:"name" json:"name"`
	Description  *string    `bson:"description" json:"description"`
	Address      *string    `bson:"address" json:"address"`
	ContactEmail *string    `bson:"contact_email" json:"contact_email"`
	ContactPhone *string    `bson:"contact_phone" json:"contact_phone"`
	Status       string     `bson:"status" json:"status"`
	CreatedAt    *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type Category struct {
	CategoryID   int32      `bson:"_id" json:"category_id"`
	Name         string     `bson:"name" json:"name"`
	Description  *string    `bson:"description" json:"description"`
	SlugCategory *string    `bson:"slug_category" json:"slug_category"`
	CreatedAt    *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type Order struct {
	OrderID    int32      `bson:"_id" json:"order_id"`
	MerchantID int32      `bson:"merchant_id" json:"merchant_id"`
	CashierID  int32      `bson:"cashier_id" json:"cashier_id"`
	TotalPrice int64      `bson:"total_price" json:"total_price"`
	CreatedAt  *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type OrderItem struct {
	OrderItemID int32      `bson:"_id" json:"order_item_id"`
	OrderID     int32      `bson:"order_id" json:"order_id"`
	ProductID   int32      `bson:"product_id" json:"product_id"`
	Quantity    int32      `bson:"quantity" json:"quantity"`
	Price       int64      `bson:"price" json:"price"`
	CreatedAt   *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type Product struct {
	ProductID    int32      `bson:"_id" json:"product_id"`
	MerchantID   int32      `bson:"merchant_id" json:"merchant_id"`
	CategoryID   int32      `bson:"category_id" json:"category_id"`
	Name         string     `bson:"name" json:"name"`
	Description  *string    `bson:"description" json:"description"`
	Price        int32      `bson:"price" json:"price"`
	CountInStock int32      `bson:"count_in_stock" json:"count_in_stock"`
	Brand        *string    `bson:"brand" json:"brand"`
	Weight       *int32     `bson:"weight" json:"weight"`
	SlugProduct  *string    `bson:"slug_product" json:"slug_product"`
	ImageProduct *string    `bson:"image_product" json:"image_product"`
	Barcode      *string    `bson:"barcode" json:"barcode"`
	CreatedAt    *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type Transaction struct {
	TransactionID int32      `bson:"_id" json:"transaction_id"`
	OrderID       int32      `bson:"order_id" json:"order_id"`
	MerchantID    int32      `bson:"merchant_id" json:"merchant_id"`
	PaymentMethod string     `bson:"payment_method" json:"payment_method"`
	Amount        int32      `bson:"amount" json:"amount"`
	ChangeAmount  *int32     `bson:"change_amount" json:"change_amount"`
	PaymentStatus *string    `bson:"payment_status" json:"payment_status"`
	CreatedAt     *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}

type MerchantDocument struct {
	DocumentID   int32      `bson:"_id" json:"document_id"`
	MerchantID   int32      `bson:"merchant_id" json:"merchant_id"`
	DocumentType string     `bson:"document_type" json:"document_type"`
	DocumentUrl  string     `bson:"document_url" json:"document_url"`
	Status       string     `bson:"status" json:"status"`
	Note         *string    `bson:"note" json:"note"`
	UploadedAt   *time.Time `bson:"uploaded_at" json:"uploaded_at"`
	CreatedAt    *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at,omitempty" json:"deleted_at"`
}
