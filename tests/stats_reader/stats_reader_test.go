// Package stats_reader_test exercises the stats-reader gRPC surface end to end:
// it seeds a real ClickHouse with known aggregates, starts the real gRPC server
// with all eleven stats services registered, and asserts the values that come
// back over the wire.
package stats_reader_test

import (
	"testing"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	orderpb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	pkgclickhouse "github.com/MamangRust/microservice-point-of-sale-pkg/clickhouse"
	statsreaderhandler "github.com/MamangRust/microservice-point-of-sale-stats-reader/handler"
	statsreaderrepo "github.com/MamangRust/microservice-point-of-sale-stats-reader/repository"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	statsYear  = 2026
	statsMonth = 1
)

type StatsReaderSuite struct {
	tests.BaseTestSuite

	chConn chDriver.Conn

	order         statspb.OrderStatsServiceClient
	orderMerchant statspb.OrderStatsByMerchantServiceClient
	orderByID     statspb.OrderStatsByIdServiceClient

	cashier         statspb.CashierStatsServiceClient
	cashierMerchant statspb.CashierStatsByMerchantServiceClient
	cashierByID     statspb.CashierStatsByIdServiceClient

	category         statspb.CategoryStatsServiceClient
	categoryMerchant statspb.CategoryStatsByMerchantServiceClient
	categoryByID     statspb.CategoryStatsByIdServiceClient

	txStatus statspb.TransactionStatsStatusServiceClient
	txMethod statspb.TransactionStatsMethodServiceClient
}

func (s *StatsReaderSuite) SetupSuite() {
	// The reader only needs ClickHouse + Redis; the context merely selects the
	// harness's primary Mongo database, which this suite does not use.
	s.Context = tests.ContextSales
	s.BaseTestSuite.SetupSuite()
	s.SetupStatsEnv()

	conn, err := pkgclickhouse.NewClient(s.Log)
	s.Require().NoError(err)
	s.Require().NoError(pkgclickhouse.ApplySchema(s.Ctx, conn, s.Log))
	s.chConn = conn

	repo := statsreaderrepo.NewClickHouseReaderRepository(conn)
	cache := statsreaderhandler.NewStatsCache(s.GetCacheStore())

	orderHandler := statsreaderhandler.NewOrderStatsHandler(repo, cache, s.Log)
	cashierHandler := statsreaderhandler.NewCashierStatsHandler(repo, cache, s.Log)
	categoryHandler := statsreaderhandler.NewCategoryStatsHandler(repo, cache, s.Log)
	transactionHandler := statsreaderhandler.NewTransactionStatsHandler(repo, cache, s.Log)

	server := grpc.NewServer()
	statspb.RegisterOrderStatsServiceServer(server, orderHandler)
	statspb.RegisterOrderStatsByMerchantServiceServer(server, orderHandler)
	statspb.RegisterOrderStatsByIdServiceServer(server, orderHandler)
	statspb.RegisterCashierStatsServiceServer(server, cashierHandler)
	statspb.RegisterCashierStatsByMerchantServiceServer(server, cashierHandler)
	statspb.RegisterCashierStatsByIdServiceServer(server, cashierHandler)
	statspb.RegisterCategoryStatsServiceServer(server, categoryHandler)
	statspb.RegisterCategoryStatsByMerchantServiceServer(server, categoryHandler)
	statspb.RegisterCategoryStatsByIdServiceServer(server, categoryHandler)
	statspb.RegisterTransactionStatsStatusServiceServer(server, transactionHandler)
	statspb.RegisterTransactionStatsMethodServiceServer(server, transactionHandler)

	client := s.GetConnection(s.RegisterServer(server))

	s.order = statspb.NewOrderStatsServiceClient(client)
	s.orderMerchant = statspb.NewOrderStatsByMerchantServiceClient(client)
	s.orderByID = statspb.NewOrderStatsByIdServiceClient(client)
	s.cashier = statspb.NewCashierStatsServiceClient(client)
	s.cashierMerchant = statspb.NewCashierStatsByMerchantServiceClient(client)
	s.cashierByID = statspb.NewCashierStatsByIdServiceClient(client)
	s.category = statspb.NewCategoryStatsServiceClient(client)
	s.categoryMerchant = statspb.NewCategoryStatsByMerchantServiceClient(client)
	s.categoryByID = statspb.NewCategoryStatsByIdServiceClient(client)
	s.txStatus = statspb.NewTransactionStatsStatusServiceClient(client)
	s.txMethod = statspb.NewTransactionStatsMethodServiceClient(client)

	s.seedClickHouse()
}

func (s *StatsReaderSuite) TearDownSuite() {
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.BaseTestSuite.TearDownSuite()
}

// seedClickHouse writes a small fixture with known aggregates:
//
//	order_daily      order 1 (two versions, FINAL keeps 1500) + order 2 (2000)
//	order_item_daily category 5 → 3 items / 500, category 6 → 3 items / 300
//	transaction_daily 2 success (3000) + 1 failed (500)
func (s *StatsReaderSuite) seedClickHouse() {
	jan := func(day int) time.Time {
		return time.Date(statsYear, time.January, day, 12, 0, 0, 0, time.UTC)
	}
	id := func(key string) uuid.UUID {
		return uuid.NewSHA1(uuid.NameSpaceOID, []byte(key))
	}

	orderInsert := `INSERT INTO order_daily
		(event_id, event_time, order_id, cashier_id, merchant_id, status, total_price, event_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	// The first two rows share a ReplacingMergeTree key; only the higher
	// event_version must survive FINAL.
	s.exec(orderInsert, id("order-1-v1"), jan(10), uint64(1), uint64(10), uint64(1), "created", int64(1000), uint64(1))
	s.exec(orderInsert, id("order-1-v1"), jan(10), uint64(1), uint64(10), uint64(1), "created", int64(1500), uint64(2))
	s.exec(orderInsert, id("order-2"), jan(11), uint64(2), uint64(20), uint64(1), "created", int64(2000), uint64(1))

	itemInsert := `INSERT INTO order_item_daily
		(event_id, event_time, order_item_id, order_id, product_id, category_id, quantity, unit_price, subtotal, event_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	s.exec(itemInsert, id("item-1"), jan(10), uint64(1), uint64(1), uint64(100), uint64(5), uint32(2), int64(100), int64(200), uint64(1))
	s.exec(itemInsert, id("item-2"), jan(10), uint64(2), uint64(1), uint64(101), uint64(5), uint32(1), int64(300), int64(300), uint64(1))
	s.exec(itemInsert, id("item-3"), jan(11), uint64(3), uint64(2), uint64(100), uint64(6), uint32(3), int64(100), int64(300), uint64(1))

	txInsert := `INSERT INTO transaction_daily
		(event_id, event_time, transaction_id, order_id, cashier_id, merchant_id, payment_method, status, amount, event_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	s.exec(txInsert, id("tx-1"), jan(10), uint64(1), uint64(1), uint64(10), uint64(1), "cash", "success", int64(1000), uint64(1))
	s.exec(txInsert, id("tx-2"), jan(11), uint64(2), uint64(2), uint64(20), uint64(1), "cash", "success", int64(2000), uint64(1))
	s.exec(txInsert, id("tx-3"), jan(11), uint64(3), uint64(2), uint64(20), uint64(1), "transfer", "failed", int64(500), uint64(1))
}

func (s *StatsReaderSuite) exec(query string, args ...any) {
	s.Require().NoError(s.chConn.Exec(s.Ctx, query, args...))
}

func (s *StatsReaderSuite) TestOrderStats() {
	monthly, err := s.order.FindMonthlyTotalRevenue(s.Ctx, &orderpb.FindYearMonthTotalRevenue{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 1)
	s.Equal(int32(2), monthly.Data[0].OrderCount)
	s.Equal(int32(3500), monthly.Data[0].TotalRevenue)
	s.Equal(int32(6), monthly.Data[0].TotalItemsSold)

	yearly, err := s.order.FindYearlyTotalRevenue(s.Ctx, &orderpb.FindYearTotalRevenue{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearly.Data, 1)
	s.Equal(int32(2), yearly.Data[0].OrderCount)
	s.Equal(int32(3500), yearly.Data[0].TotalRevenue)
	s.Equal(int32(6), yearly.Data[0].TotalItemsSold)
	s.Equal(int32(2), yearly.Data[0].ActiveCashiers)
	s.Equal(int32(2), yearly.Data[0].UniqueProductsSold)

	// order 1 alone must report the surviving ReplacingMergeTree version.
	byOrder, err := s.orderByID.FindYearlyTotalRevenueById(s.Ctx, &orderpb.FindYearTotalRevenueById{
		Year: statsYear, OrderId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(byOrder.Data, 1)
	s.Equal(int32(1500), byOrder.Data[0].TotalRevenue, "FINAL must keep the highest event_version")
	s.Equal(int32(3), byOrder.Data[0].TotalItemsSold)
}

func (s *StatsReaderSuite) TestOrderStatsByMerchant() {
	resp, err := s.orderMerchant.FindMonthlyTotalRevenueByMerchant(s.Ctx, &orderpb.FindYearMonthTotalRevenueByMerchant{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(resp.Data, 1)
	s.Equal(int32(3500), resp.Data[0].TotalRevenue)
}

func (s *StatsReaderSuite) TestCashierStats() {
	resp, err := s.cashier.FindMonthSales(s.Ctx, &cashierpb.FindYearCashier{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(resp.Data, 2)

	byCashier := map[int32]*cashierpb.CashierResponseMonthSales{}
	for _, row := range resp.Data {
		byCashier[row.CashierId] = row
	}
	s.Equal(int32(1), byCashier[10].OrderCount)
	s.Equal(int32(1500), byCashier[10].TotalSales)
	s.Equal(int32(1), byCashier[20].OrderCount)
	s.Equal(int32(2000), byCashier[20].TotalSales)

	byID, err := s.cashierByID.FindMonthSalesById(s.Ctx, &cashierpb.FindYearCashierById{
		CashierId: 20, Year: statsYear,
	})
	s.Require().NoError(err)
	s.Require().Len(byID.Data, 1)
	s.Equal(int32(2000), byID.Data[0].TotalSales)
}

func (s *StatsReaderSuite) TestCategoryStats() {
	resp, err := s.category.FindMonthPrice(s.Ctx, &categorypb.FindYearCategory{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(resp.Data, 2)

	byCategory := map[int32]*categorypb.CategoryMonthPriceResponse{}
	for _, row := range resp.Data {
		byCategory[row.CategoryId] = row
	}
	s.Equal(int32(1), byCategory[5].OrderCount)
	s.Equal(int32(3), byCategory[5].ItemsSold)
	s.Equal(int32(500), byCategory[5].TotalRevenue)
	s.Equal(int32(3), byCategory[6].ItemsSold)
	s.Equal(int32(300), byCategory[6].TotalRevenue)

	monthly, err := s.category.FindMonthlyTotalPrices(s.Ctx, &categorypb.FindYearMonthTotalPrices{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 1)
	s.Equal(int32(800), monthly.Data[0].TotalRevenue)

	byID, err := s.categoryByID.FindYearPriceById(s.Ctx, &categorypb.FindYearCategoryById{
		CategoryId: 5, Year: statsYear,
	})
	s.Require().NoError(err)
	s.Require().Len(byID.Data, 1)
	s.Equal(int32(3), byID.Data[0].ItemsSold)
	s.Equal(int32(500), byID.Data[0].TotalRevenue)
	s.Equal(int32(2), byID.Data[0].UniqueProductsSold)
}

func (s *StatsReaderSuite) TestTransactionStats() {
	success, err := s.txStatus.FindMonthStatusSuccess(s.Ctx, &transactionpb.FindMonthlyTransactionStatus{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(success.Data, 1)
	s.Equal(int32(2), success.Data[0].TotalSuccess)
	s.Equal(int32(3000), success.Data[0].TotalAmount)

	failed, err := s.txStatus.FindMonthStatusFailed(s.Ctx, &transactionpb.FindMonthlyTransactionStatus{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(failed.Data, 1)
	s.Equal(int32(1), failed.Data[0].TotalFailed)
	s.Equal(int32(500), failed.Data[0].TotalAmount)

	methodSuccess, err := s.txMethod.FindMonthMethodSuccess(s.Ctx, &transactionpb.MonthTransactionMethod{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(methodSuccess.Data, 1)
	s.Equal("cash", methodSuccess.Data[0].PaymentMethod)
	s.Equal(int32(2), methodSuccess.Data[0].TotalTransactions)
	s.Equal(int32(3000), methodSuccess.Data[0].TotalAmount)

	methodFailed, err := s.txMethod.FindMonthMethodFailed(s.Ctx, &transactionpb.MonthTransactionMethod{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(methodFailed.Data, 1)
	s.Equal("transfer", methodFailed.Data[0].PaymentMethod)
	s.Equal(int32(1), methodFailed.Data[0].TotalTransactions)
	s.Equal(int32(500), methodFailed.Data[0].TotalAmount)
}

// TestAllRegisteredServicesReachable proves every one of the eleven services is
// wired to a handler: a missing registration would answer Unimplemented.
func (s *StatsReaderSuite) TestAllRegisteredServicesReachable() {
	cases := []struct {
		service string
		call    func() error
	}{
		{"OrderStatsService", func() error {
			_, err := s.order.FindYearlyRevenue(s.Ctx, &orderpb.FindYearOrder{Year: statsYear})
			return err
		}},
		{"OrderStatsByMerchantService", func() error {
			_, err := s.orderMerchant.FindYearlyRevenueByMerchant(s.Ctx, &orderpb.FindYearOrderByMerchant{Year: statsYear, MerchantId: 1})
			return err
		}},
		{"OrderStatsByIdService", func() error {
			_, err := s.orderByID.FindMonthlyTotalRevenueById(s.Ctx, &orderpb.FindYearMonthTotalRevenueById{Year: statsYear, Month: statsMonth, OrderId: 1})
			return err
		}},
		{"CashierStatsService", func() error {
			_, err := s.cashier.FindYearlyTotalSales(s.Ctx, &cashierpb.FindYearTotalSales{Year: statsYear})
			return err
		}},
		{"CashierStatsByMerchantService", func() error {
			_, err := s.cashierMerchant.FindYearlyTotalSalesByMerchant(s.Ctx, &cashierpb.FindYearTotalSalesByMerchant{Year: statsYear, MerchantId: 1})
			return err
		}},
		{"CashierStatsByIdService", func() error {
			_, err := s.cashierByID.FindYearlyTotalSalesById(s.Ctx, &cashierpb.FindYearTotalSalesById{Year: statsYear, CashierId: 10})
			return err
		}},
		{"CategoryStatsService", func() error {
			_, err := s.category.FindYearlyTotalPrices(s.Ctx, &categorypb.FindYearTotalPrices{Year: statsYear})
			return err
		}},
		{"CategoryStatsByMerchantService", func() error {
			_, err := s.categoryMerchant.FindYearlyTotalPricesByMerchant(s.Ctx, &categorypb.FindYearTotalPriceByMerchant{Year: statsYear, MerchantId: 1})
			return err
		}},
		{"CategoryStatsByIdService", func() error {
			_, err := s.categoryByID.FindYearlyTotalPricesById(s.Ctx, &categorypb.FindYearTotalPriceById{Year: statsYear, CategoryId: 5})
			return err
		}},
		{"TransactionStatsStatusService", func() error {
			_, err := s.txStatus.FindYearStatusSuccess(s.Ctx, &transactionpb.FindYearlyTransactionStatus{Year: statsYear})
			return err
		}},
		{"TransactionStatsMethodService", func() error {
			_, err := s.txMethod.FindYearMethodSuccess(s.Ctx, &transactionpb.YearTransactionMethod{Year: statsYear})
			return err
		}},
	}

	for _, tc := range cases {
		s.Run(tc.service, func() {
			err := tc.call()
			s.NoError(err)
			s.NotEqual(codes.Unimplemented, status.Code(err), "%s is not registered", tc.service)
		})
	}
}

func TestStatsReaderSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	suite.Run(t, new(StatsReaderSuite))
}
