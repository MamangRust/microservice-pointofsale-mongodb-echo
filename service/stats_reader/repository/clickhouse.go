package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseReaderRepository struct {
	conn clickhouse.Conn
}

func NewClickHouseReaderRepository(conn clickhouse.Conn) *ClickHouseReaderRepository {
	return &ClickHouseReaderRepository{conn: conn}
}

// ── Filter helpers ──────────────────────────────────────────────────────────
//
// Zero-valued filter fields mean "no filter". order_daily carries merchant_id;
// order_item_daily does not, so its ByMerchant variants filter via subquery on
// order_daily.

type orderFilter struct {
	year, month, merchantID, cashierID, orderID int
}

type itemFilter struct {
	year, month, merchantID, categoryID, orderID int
}

type transactionFilter struct {
	year, month, merchantID int
}

func orderWhere(f orderFilter) (string, []any) {
	var conds []string
	var args []any
	if f.year > 0 {
		conds = append(conds, "toYear(event_time) = ?")
		args = append(args, f.year)
	}
	if f.month > 0 {
		conds = append(conds, "toMonth(event_time) = ?")
		args = append(args, f.month)
	}
	if f.merchantID > 0 {
		conds = append(conds, "merchant_id = ?")
		args = append(args, f.merchantID)
	}
	if f.cashierID > 0 {
		conds = append(conds, "cashier_id = ?")
		args = append(args, f.cashierID)
	}
	if f.orderID > 0 {
		conds = append(conds, "order_id = ?")
		args = append(args, f.orderID)
	}
	return strings.Join(conds, " AND "), args
}

func itemWhere(f itemFilter) (string, []any) {
	var conds []string
	var args []any
	if f.year > 0 {
		conds = append(conds, "toYear(event_time) = ?")
		args = append(args, f.year)
	}
	if f.month > 0 {
		conds = append(conds, "toMonth(event_time) = ?")
		args = append(args, f.month)
	}
	if f.merchantID > 0 {
		conds = append(conds, "order_id IN (SELECT order_id FROM order_daily FINAL WHERE merchant_id = ?)")
		args = append(args, f.merchantID)
	}
	if f.categoryID > 0 {
		conds = append(conds, "category_id = ?")
		args = append(args, f.categoryID)
	}
	if f.orderID > 0 {
		conds = append(conds, "order_id = ?")
		args = append(args, f.orderID)
	}
	return strings.Join(conds, " AND "), args
}

func transactionWhere(f transactionFilter) (string, []any) {
	var conds []string
	var args []any
	if f.year > 0 {
		conds = append(conds, "toYear(event_time) = ?")
		args = append(args, f.year)
	}
	if f.month > 0 {
		conds = append(conds, "toMonth(event_time) = ?")
		args = append(args, f.month)
	}
	if f.merchantID > 0 {
		conds = append(conds, "merchant_id = ?")
		args = append(args, f.merchantID)
	}
	return strings.Join(conds, " AND "), args
}

// POS stores both "success" (API creates) and "Completed"/"completed" (legacy
// seeder data) for a successful payment; match case-insensitively so the reader
// cross-checks with the OLTP payment_status values. Anything else counts as
// failed.
const successStatus = "lower(status) IN ('success', 'completed')"

func (r *ClickHouseReaderRepository) query(ctx context.Context, name, query string, args []any, scan func(rows driver.Rows) error) error {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query %s: %w", name, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("scan %s: %w", name, err)
		}
	}
	return rows.Err()
}
