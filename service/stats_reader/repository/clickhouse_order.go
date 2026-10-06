package repository

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ── Order stats ─────────────────────────────────────────────────────────────
//
// order_daily holds revenue/order/cashier aggregates; total_items_sold and
// unique_products_sold only exist in order_item_daily, so they are aggregated
// separately and merged in Go on the month/year key.

type orderAggRow struct {
	Year           string
	Month          string
	OrderCount     uint64
	TotalRevenue   int64
	ActiveCashiers uint64
}

type itemAggRow struct {
	ItemsSold      uint64
	UniqueProducts uint64
}

func (r *ClickHouseReaderRepository) orderAggregate(ctx context.Context, f orderFilter, byMonth bool) ([]orderAggRow, error) {
	where, args := orderWhere(f)
	keyExpr, orderExpr, groupExtra := "''", "year", ""
	if byMonth {
		keyExpr, orderExpr, groupExtra = "formatDateTime(event_time, '%b')", "toMonth(event_time)", ", toMonth(event_time)"
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(event_time)) AS year,
		       %s AS month_key,
		       count() AS order_count,
		       sum(total_price) AS total_revenue,
		       uniqExact(cashier_id) AS active_cashiers
		FROM order_daily FINAL
		WHERE %s
		GROUP BY year, month_key%s
		ORDER BY %s
	`, keyExpr, where, groupExtra, orderExpr)

	var results []orderAggRow
	err := r.query(ctx, "order aggregate", query, args, func(rows driver.Rows) error {
		var row orderAggRow
		if err := rows.Scan(&row.Year, &row.Month, &row.OrderCount, &row.TotalRevenue, &row.ActiveCashiers); err != nil {
			return err
		}
		results = append(results, row)
		return nil
	})
	return results, err
}

func (r *ClickHouseReaderRepository) itemAggregate(ctx context.Context, f itemFilter, byMonth bool) (map[string]itemAggRow, error) {
	where, args := itemWhere(f)
	// ORDER BY must reference the SELECT alias (item_key): the year expression is
	// aliased AS item_key and there is no bare "year" column in scope.
	keyExpr, orderExpr, groupExtra := "toString(toYear(event_time))", "item_key", ""
	if byMonth {
		keyExpr, orderExpr, groupExtra = "formatDateTime(event_time, '%b')", "toMonth(event_time)", ", toMonth(event_time)"
	}
	query := fmt.Sprintf(`
		SELECT %s AS item_key,
		       sum(quantity) AS items_sold,
		       uniqExact(product_id) AS unique_products
		FROM order_item_daily FINAL
		WHERE %s
		GROUP BY item_key%s
		ORDER BY %s
	`, keyExpr, where, groupExtra, orderExpr)

	items := map[string]itemAggRow{}
	err := r.query(ctx, "item aggregate", query, args, func(rows driver.Rows) error {
		var key string
		var agg itemAggRow
		if err := rows.Scan(&key, &agg.ItemsSold, &agg.UniqueProducts); err != nil {
			return err
		}
		items[key] = agg
		return nil
	})
	return items, err
}

func mergeOrderMonthlyTotalRevenue(rows []orderAggRow, items map[string]itemAggRow) []OrderMonthlyTotalRevenue {
	out := make([]OrderMonthlyTotalRevenue, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrderMonthlyTotalRevenue{
			Year:           row.Year,
			Month:          row.Month,
			OrderCount:     row.OrderCount,
			TotalRevenue:   row.TotalRevenue,
			TotalItemsSold: items[row.Month].ItemsSold,
		})
	}
	return out
}

func mergeOrderYearly(rows []orderAggRow, items map[string]itemAggRow) []OrderYearlyTotalRevenue {
	out := make([]OrderYearlyTotalRevenue, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrderYearlyTotalRevenue{
			Year:               row.Year,
			OrderCount:         row.OrderCount,
			TotalRevenue:       row.TotalRevenue,
			TotalItemsSold:     items[row.Year].ItemsSold,
			ActiveCashiers:     row.ActiveCashiers,
			UniqueProductsSold: items[row.Year].UniqueProducts,
		})
	}
	return out
}

func (r *ClickHouseReaderRepository) GetMonthlyTotalRevenue(ctx context.Context, year, month int) ([]OrderMonthlyTotalRevenue, error) {
	f := orderFilter{year: year, month: month}
	rows, err := r.orderAggregate(ctx, f, true)
	if err != nil {
		return nil, err
	}
	items, err := r.itemAggregate(ctx, itemFilter{year: year, month: month}, true)
	if err != nil {
		return nil, err
	}
	return mergeOrderMonthlyTotalRevenue(rows, items), nil
}

func (r *ClickHouseReaderRepository) GetYearlyTotalRevenue(ctx context.Context, year int) ([]OrderYearlyTotalRevenue, error) {
	return r.orderYearlyTotalRevenue(ctx, orderFilter{year: year})
}

// orderYearlyTotalRevenue backs both FindYearlyTotalRevenue and
// FindYearlyRevenue — their pb responses carry the same aggregates.
func (r *ClickHouseReaderRepository) orderYearlyTotalRevenue(ctx context.Context, f orderFilter) ([]OrderYearlyTotalRevenue, error) {
	rows, err := r.orderAggregate(ctx, f, false)
	if err != nil {
		return nil, err
	}
	items, err := r.itemAggregate(ctx, itemFilter{year: f.year, merchantID: f.merchantID, orderID: f.orderID}, false)
	if err != nil {
		return nil, err
	}
	return mergeOrderYearly(rows, items), nil
}

func (r *ClickHouseReaderRepository) GetMonthlyRevenue(ctx context.Context, year int) ([]OrderMonthly, error) {
	f := orderFilter{year: year}
	rows, err := r.orderAggregate(ctx, f, true)
	if err != nil {
		return nil, err
	}
	items, err := r.itemAggregate(ctx, itemFilter{year: year}, true)
	if err != nil {
		return nil, err
	}
	out := make([]OrderMonthly, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrderMonthly{
			Month:          row.Month,
			OrderCount:     row.OrderCount,
			TotalRevenue:   row.TotalRevenue,
			TotalItemsSold: items[row.Month].ItemsSold,
		})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetYearlyRevenue(ctx context.Context, year int) ([]OrderYearly, error) {
	rows, err := r.orderYearlyTotalRevenue(ctx, orderFilter{year: year})
	if err != nil {
		return nil, err
	}
	out := make([]OrderYearly, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrderYearly{
			Year:               row.Year,
			OrderCount:         row.OrderCount,
			TotalRevenue:       row.TotalRevenue,
			TotalItemsSold:     row.TotalItemsSold,
			ActiveCashiers:     row.ActiveCashiers,
			UniqueProductsSold: row.UniqueProductsSold,
		})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetMonthlyTotalRevenueByMerchant(ctx context.Context, year, month, merchantID int) ([]OrderMonthlyTotalRevenue, error) {
	f := orderFilter{year: year, month: month, merchantID: merchantID}
	rows, err := r.orderAggregate(ctx, f, true)
	if err != nil {
		return nil, err
	}
	items, err := r.itemAggregate(ctx, itemFilter{year: year, month: month, merchantID: merchantID}, true)
	if err != nil {
		return nil, err
	}
	return mergeOrderMonthlyTotalRevenue(rows, items), nil
}

func (r *ClickHouseReaderRepository) GetYearlyTotalRevenueByMerchant(ctx context.Context, year, merchantID int) ([]OrderYearlyTotalRevenue, error) {
	return r.orderYearlyTotalRevenue(ctx, orderFilter{year: year, merchantID: merchantID})
}

func (r *ClickHouseReaderRepository) GetMonthlyRevenueByMerchant(ctx context.Context, year, merchantID int) ([]OrderMonthly, error) {
	f := orderFilter{year: year, merchantID: merchantID}
	rows, err := r.orderAggregate(ctx, f, true)
	if err != nil {
		return nil, err
	}
	items, err := r.itemAggregate(ctx, itemFilter{year: year, merchantID: merchantID}, true)
	if err != nil {
		return nil, err
	}
	out := make([]OrderMonthly, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrderMonthly{
			Month:          row.Month,
			OrderCount:     row.OrderCount,
			TotalRevenue:   row.TotalRevenue,
			TotalItemsSold: items[row.Month].ItemsSold,
		})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetYearlyRevenueByMerchant(ctx context.Context, year, merchantID int) ([]OrderYearly, error) {
	rows, err := r.orderYearlyTotalRevenue(ctx, orderFilter{year: year, merchantID: merchantID})
	if err != nil {
		return nil, err
	}
	out := make([]OrderYearly, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrderYearly{
			Year:               row.Year,
			OrderCount:         row.OrderCount,
			TotalRevenue:       row.TotalRevenue,
			TotalItemsSold:     row.TotalItemsSold,
			ActiveCashiers:     row.ActiveCashiers,
			UniqueProductsSold: row.UniqueProductsSold,
		})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetMonthlyTotalRevenueById(ctx context.Context, year, month, orderID int) ([]OrderMonthlyTotalRevenue, error) {
	f := orderFilter{year: year, month: month, orderID: orderID}
	rows, err := r.orderAggregate(ctx, f, true)
	if err != nil {
		return nil, err
	}
	items, err := r.itemAggregate(ctx, itemFilter{year: year, month: month, orderID: orderID}, true)
	if err != nil {
		return nil, err
	}
	return mergeOrderMonthlyTotalRevenue(rows, items), nil
}

func (r *ClickHouseReaderRepository) GetYearlyTotalRevenueById(ctx context.Context, year, orderID int) ([]OrderYearlyTotalRevenue, error) {
	return r.orderYearlyTotalRevenue(ctx, orderFilter{year: year, orderID: orderID})
}
