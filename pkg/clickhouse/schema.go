package clickhouse

import (
	"context"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// schemaSQL mirrors pkg/clickhouse/schema.sql. It is embedded here so
// stats-writer can guarantee the tables exist before consuming/backfilling
// without depending on the container init script.
const schemaSQL = `
CREATE TABLE IF NOT EXISTS order_daily
(
    event_id      UUID,
    event_time    DateTime,
    order_id      UInt64,
    cashier_id    UInt64,
    merchant_id   UInt64,
    status        LowCardinality(String),
    total_price   Int64,
    event_version UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (toDate(event_time), order_id, event_id);

CREATE TABLE IF NOT EXISTS order_item_daily
(
    event_id      UUID,
    event_time    DateTime,
    order_item_id UInt64,
    order_id      UInt64,
    product_id    UInt64,
    category_id   UInt64,
    quantity      UInt32,
    unit_price    Int64,
    subtotal      Int64,
    event_version UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (toDate(event_time), order_item_id, event_id);

CREATE TABLE IF NOT EXISTS transaction_daily
(
    event_id       UUID,
    event_time     DateTime,
    transaction_id UInt64,
    order_id       UInt64,
    cashier_id     UInt64,
    merchant_id    UInt64,
    payment_method LowCardinality(String),
    status         LowCardinality(String),
    amount         Int64,
    event_version  UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (toDate(event_time), transaction_id, event_id);

CREATE TABLE IF NOT EXISTS cashier_daily
(
    event_id      UUID,
    event_time    DateTime,
    cashier_id    UInt64,
    merchant_id   UInt64,
    status        LowCardinality(String),
    event_version UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (toDate(event_time), cashier_id, event_id);
`

// ApplySchema executes the stats table DDL inside the configured database
// (CLICKHOUSE_DATABASE), creating it first when it is not the default. Every
// statement is CREATE TABLE IF NOT EXISTS, so applying it repeatedly is a
// no-op. Using a dedicated database keeps the POS stats tables isolated from
// the payment-gateway/ecommerce pipelines when all share one ClickHouse server.
func ApplySchema(ctx context.Context, conn clickhouse.Conn, log logger.LoggerInterface) error {
	dbName := viper.GetString("CLICKHOUSE_DATABASE")
	if dbName != "" && dbName != "default" {
		if err := conn.Exec(ctx, "CREATE DATABASE IF NOT EXISTS "+dbName); err != nil {
			log.Error("Failed to create ClickHouse database", zap.Error(err), zap.String("database", dbName))
			return err
		}
	}
	for _, stmt := range strings.Split(schemaSQL, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := conn.Exec(ctx, stmt); err != nil {
			log.Error("Failed to apply ClickHouse schema statement", zap.Error(err), zap.String("statement", stmt))
			return err
		}
	}
	return nil
}
