package repository

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ── Category stats ──────────────────────────────────────────────────────────

type categoryAggRow struct {
	Year           string
	Month          string
	CategoryID     uint64
	OrderCount     uint64
	ItemsSold      uint64
	TotalRevenue   int64
	UniqueProducts uint64
}

// categoryAggregate groups order_item_daily by month and/or category. When
// byCategory is false the category_id/unique_products slots are constants.
func (r *ClickHouseReaderRepository) categoryAggregate(ctx context.Context, f itemFilter, byMonth, byCategory bool) ([]categoryAggRow, error) {
	where, args := itemWhere(f)
	keyExpr, orderExpr, groupExtra := "''", "year", ""
	if byMonth {
		keyExpr, orderExpr, groupExtra = "formatDateTime(event_time, '%b')", "toMonth(event_time)", ", toMonth(event_time)"
	}
	categoryExpr, productExpr := "toUInt64(0)", "toUInt64(0)"
	if byCategory {
		categoryExpr, productExpr = "category_id", "uniqExact(product_id)"
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(event_time)) AS year,
		       %s AS month_key,
		       %s AS category_id,
		       uniqExact(order_id) AS order_count,
		       sum(quantity) AS items_sold,
		       sum(subtotal) AS total_revenue,
		       %s AS unique_products
		FROM order_item_daily FINAL
		WHERE %s
		GROUP BY year, month_key, category_id%s
		ORDER BY %s, total_revenue DESC
	`, keyExpr, categoryExpr, productExpr, where, groupExtra, orderExpr)

	var results []categoryAggRow
	err := r.query(ctx, "category aggregate", query, args, func(rows driver.Rows) error {
		var row categoryAggRow
		if err := rows.Scan(&row.Year, &row.Month, &row.CategoryID, &row.OrderCount, &row.ItemsSold, &row.TotalRevenue, &row.UniqueProducts); err != nil {
			return err
		}
		results = append(results, row)
		return nil
	})
	return results, err
}

func (r *ClickHouseReaderRepository) GetMonthlyTotalPrices(ctx context.Context, year, month int) ([]CategoryMonthTotalPrice, error) {
	rows, err := r.categoryAggregate(ctx, itemFilter{year: year, month: month}, true, false)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryMonthTotalPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryMonthTotalPrice{Year: row.Year, Month: row.Month, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetYearlyTotalPrices(ctx context.Context, year int) ([]CategoryYearTotalPrice, error) {
	rows, err := r.categoryAggregate(ctx, itemFilter{year: year}, false, false)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryYearTotalPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryYearTotalPrice{Year: row.Year, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

// GetMonthPrice returns per-category monthly price aggregates. category_name is
// not stored in ClickHouse; the front end resolves it via the category domain
// service.
func (r *ClickHouseReaderRepository) GetMonthPrice(ctx context.Context, year int) ([]CategoryMonthPrice, error) {
	return categoryMonthPrice(r.categoryAggregate(ctx, itemFilter{year: year}, true, true))
}

func (r *ClickHouseReaderRepository) GetYearPrice(ctx context.Context, year int) ([]CategoryYearPrice, error) {
	return categoryYearPrice(r.categoryAggregate(ctx, itemFilter{year: year}, false, true))
}

func (r *ClickHouseReaderRepository) GetMonthlyTotalPricesByMerchant(ctx context.Context, year, month, merchantID int) ([]CategoryMonthTotalPrice, error) {
	rows, err := r.categoryAggregate(ctx, itemFilter{year: year, month: month, merchantID: merchantID}, true, false)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryMonthTotalPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryMonthTotalPrice{Year: row.Year, Month: row.Month, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetYearlyTotalPricesByMerchant(ctx context.Context, year, merchantID int) ([]CategoryYearTotalPrice, error) {
	rows, err := r.categoryAggregate(ctx, itemFilter{year: year, merchantID: merchantID}, false, false)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryYearTotalPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryYearTotalPrice{Year: row.Year, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetMonthPriceByMerchant(ctx context.Context, year, merchantID int) ([]CategoryMonthPrice, error) {
	return categoryMonthPrice(r.categoryAggregate(ctx, itemFilter{year: year, merchantID: merchantID}, true, true))
}

func (r *ClickHouseReaderRepository) GetYearPriceByMerchant(ctx context.Context, year, merchantID int) ([]CategoryYearPrice, error) {
	return categoryYearPrice(r.categoryAggregate(ctx, itemFilter{year: year, merchantID: merchantID}, false, true))
}

func (r *ClickHouseReaderRepository) GetMonthlyTotalPricesById(ctx context.Context, year, month, categoryID int) ([]CategoryMonthTotalPrice, error) {
	rows, err := r.categoryAggregate(ctx, itemFilter{year: year, month: month, categoryID: categoryID}, true, false)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryMonthTotalPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryMonthTotalPrice{Year: row.Year, Month: row.Month, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetYearlyTotalPricesById(ctx context.Context, year, categoryID int) ([]CategoryYearTotalPrice, error) {
	rows, err := r.categoryAggregate(ctx, itemFilter{year: year, categoryID: categoryID}, false, false)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryYearTotalPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryYearTotalPrice{Year: row.Year, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

func (r *ClickHouseReaderRepository) GetMonthPriceById(ctx context.Context, year, categoryID int) ([]CategoryMonthPrice, error) {
	return categoryMonthPrice(r.categoryAggregate(ctx, itemFilter{year: year, categoryID: categoryID}, true, true))
}

func (r *ClickHouseReaderRepository) GetYearPriceById(ctx context.Context, year, categoryID int) ([]CategoryYearPrice, error) {
	return categoryYearPrice(r.categoryAggregate(ctx, itemFilter{year: year, categoryID: categoryID}, false, true))
}

func categoryMonthPrice(rows []categoryAggRow, err error) ([]CategoryMonthPrice, error) {
	if err != nil {
		return nil, err
	}
	out := make([]CategoryMonthPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryMonthPrice{Month: row.Month, CategoryID: row.CategoryID, OrderCount: row.OrderCount, ItemsSold: row.ItemsSold, TotalRevenue: row.TotalRevenue})
	}
	return out, nil
}

func categoryYearPrice(rows []categoryAggRow, err error) ([]CategoryYearPrice, error) {
	if err != nil {
		return nil, err
	}
	out := make([]CategoryYearPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryYearPrice{Year: row.Year, CategoryID: row.CategoryID, OrderCount: row.OrderCount, ItemsSold: row.ItemsSold, TotalRevenue: row.TotalRevenue, UniqueProductsSold: row.UniqueProducts})
	}
	return out, nil
}
