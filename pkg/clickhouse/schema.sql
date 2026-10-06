-- ClickHouse Schema for POS Stats (F3)
--
-- The stats pipeline materializes OLTP events into ClickHouse:
--   domain services --(outbox)--> Kafka stats.pos.<domain>.event --> stats-writer --> ClickHouse
--   apigateway --(gRPC)--> stats-reader --> ClickHouse
--
-- ReplacingMergeTree(event_version) dedupes at-least-once redeliveries: a row
-- with the same ORDER BY key and a newer event_version replaces the older one.
-- backfill writes event_version = unix timestamp of the run so a re-backfill
-- supersedes prior rows instead of duplicating aggregates.

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
