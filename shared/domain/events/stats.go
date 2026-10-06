// Package events defines the domain event payloads used by the stats pipeline
// (F3). Producers publish a StatsEnvelope on a stats.pos.<domain>.event topic;
// stats-writer unwraps it, dedups on EventID, and materializes the payload
// into ClickHouse.
package events

import "encoding/json"

// StatsEnvelope is the standard envelope for stats events. EventID is the
// idempotency key: stats-writer skips events whose EventID was already
// consumed, and ClickHouse keeps it as part of the row's primary key so a
// redelivery cannot duplicate aggregates.
type StatsEnvelope struct {
	EventID string          `json:"event_id"`
	Payload json.RawMessage `json:"payload"`
}

// OrderEvent is published when an order is created or its status changes.
// It feeds the order revenue statistics. TotalPrice matches the POS orders
// column (BIGINT).
type OrderEvent struct {
	OrderID    int32  `json:"order_id"`
	CashierID  int32  `json:"cashier_id"`
	MerchantID int32  `json:"merchant_id"`
	TotalPrice int64  `json:"total_price"`
	Status     string `json:"status"`
	EventTime  string `json:"event_time"`
}

// OrderItemEvent is published when an order item is created. CategoryID is
// denormalized here so the category/product sold statistics can be served
// purely from ClickHouse without a cross-service DB join.
type OrderItemEvent struct {
	OrderItemID int32  `json:"order_item_id"`
	OrderID     int32  `json:"order_id"`
	ProductID   int32  `json:"product_id"`
	CategoryID  int32  `json:"category_id"`
	Quantity    int32  `json:"quantity"`
	UnitPrice   int32  `json:"unit_price"`
	Subtotal    int32  `json:"subtotal"`
	EventTime   string `json:"event_time"`
}

// TransactionEvent is published when a transaction is created or its status
// changes. It feeds the transaction success/failed and payment-method
// statistics.
type TransactionEvent struct {
	TransactionID int32  `json:"transaction_id"`
	OrderID       int32  `json:"order_id"`
	CashierID     int32  `json:"cashier_id"`
	MerchantID    int32  `json:"merchant_id"`
	PaymentMethod string `json:"payment_method"`
	Amount        int32  `json:"amount"`
	Status        string `json:"status"`
	EventTime     string `json:"event_time"`
}

// ProductEvent is published when a product is created or its price changes.
type ProductEvent struct {
	ProductID  int32  `json:"product_id"`
	CategoryID int32  `json:"category_id"`
	Price      int32  `json:"price"`
	EventTime  string `json:"event_time"`
}

// CashierEvent is published when a cashier is created or their status changes.
type CashierEvent struct {
	CashierID  int32  `json:"cashier_id"`
	MerchantID int32  `json:"merchant_id"`
	Status     string `json:"status"`
	EventTime  string `json:"event_time"`
}
