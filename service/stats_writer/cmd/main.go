package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	transactionadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/clickhouse"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/kafka"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/backfill"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/handler"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/repository"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/usecase"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	if err := dotenv.Viper(); err != nil {
		zap.L().Error("Failed to load configuration", zap.Error(err))
	}
	log, _ := logger.NewLogger("stats-writer", nil)

	// The ClickHouse database must exist before NewClient can ping (it connects
	// with the configured database as default), so create it first.
	if err := clickhouse.EnsureDatabase(log); err != nil {
		log.Fatal("Failed to ensure ClickHouse database", zap.Error(err))
	}
	chConn, err := clickhouse.NewClient(log)
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}

	// Guarantee the stats tables exist before any read/write.
	if err := clickhouse.ApplySchema(context.Background(), chConn, log); err != nil {
		log.Fatal("Failed to apply ClickHouse schema", zap.Error(err))
	}

	repo := repository.NewClickhouseRepository(chConn, log)
	uc := usecase.NewStatsUseCase(repo)

	// `stats-writer backfill` materializes historical OLTP rows into
	// ClickHouse and exits. Everything else runs the live Kafka consumer.
	//
	// Backfill reads every source through the service that owns it (one owner
	// per collection): orders from order, transactions from transaction, order
	// items from order_item, and products from product. Each dependency gets its
	// own guard so a slow service fails fast instead of stalling the walk.
	if len(os.Args) > 1 && os.Args[1] == "backfill" {
		orderAddr := getEnv("GRPC_ORDER_ADDR", "localhost:50058")
		orderConn, err := server.NewGRPCClient(orderAddr)
		if err != nil {
			log.Fatal("Failed to connect to order service", zap.String("addr", orderAddr), zap.Error(err))
		}
		defer orderConn.Close()

		guardOrder := resilience.NewDependencyGuard("order", 5, 30, 100, 3*time.Second, log)
		orderRepo := orderadapter.New(
			pborder.NewOrderServiceClient(orderConn),
			adapter.WithDependencyGuard(guardOrder),
		)

		productAddr := getEnv("GRPC_PRODUCT_ADDR", "localhost:50059")
		productConn, err := server.NewGRPCClient(productAddr)
		if err != nil {
			log.Fatal("Failed to connect to product service", zap.String("addr", productAddr), zap.Error(err))
		}
		defer productConn.Close()

		guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, log)
		productRepo := productadapter.New(
			pbproduct.NewProductServiceClient(productConn),
			adapter.WithDependencyGuard(guardProduct),
		)

		transactionAddr := getEnv("GRPC_TRANSACTION_ADDR", "localhost:50060")
		transactionConn, err := server.NewGRPCClient(transactionAddr)
		if err != nil {
			log.Fatal("Failed to connect to transaction service", zap.String("addr", transactionAddr), zap.Error(err))
		}
		defer transactionConn.Close()

		guardTransaction := resilience.NewDependencyGuard("transaction", 5, 30, 100, 3*time.Second, log)
		transactionRepo := transactionadapter.New(
			pbtransaction.NewTransactionServiceClient(transactionConn),
			adapter.WithDependencyGuard(guardTransaction),
		)

		orderItemAddr := getEnv("GRPC_ORDERITEM_ADDR", "localhost:50057")
		orderItemConn, err := server.NewGRPCClient(orderItemAddr)
		if err != nil {
			log.Fatal("Failed to connect to order_item service", zap.String("addr", orderItemAddr), zap.Error(err))
		}
		defer orderItemConn.Close()

		guardOrderItem := resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, log)
		orderItemRepo := orderitemadapter.New(
			pborderitem.NewOrderItemServiceClient(orderItemConn),
			adapter.WithDependencyGuard(guardOrderItem),
		)

		cashierAddr := getEnv("GRPC_CASHIER_ADDR", "localhost:50055")
		cashierConn, err := server.NewGRPCClient(cashierAddr)
		if err != nil {
			log.Fatal("Failed to connect to cashier service", zap.String("addr", cashierAddr), zap.Error(err))
		}
		defer cashierConn.Close()

		guardCashier := resilience.NewDependencyGuard("cashier", 5, 30, 100, 3*time.Second, log)
		cashierRepo := cashieradapter.New(
			pbcashier.NewCashierServiceClient(cashierConn),
			adapter.WithDependencyGuard(guardCashier),
		)

		bf := backfill.New(log, repo, orderRepo, orderItemRepo, productRepo, transactionRepo, cashierRepo)

		if err := bf.Run(context.Background()); err != nil {
			log.Fatal("Backfill failed", zap.Error(err))
		}
		log.Info("backfill finished",
			zap.String("order_service", orderAddr),
			zap.String("product_service", productAddr),
			zap.String("transaction_service", transactionAddr),
			zap.String("order_item_service", orderItemAddr),
			zap.String("cashier_service", cashierAddr),
		)
		return
	}

	brokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka:9092", "localhost:9092"}
	}
	k := kafka.NewKafka(log, brokers)

	statsHandler := handler.NewStatsHandler(uc, log)
	if err := k.StartConsumers(handler.StatsTopics(), "pos-stats-writer", statsHandler); err != nil {
		log.Fatal("Failed to start Kafka consumers", zap.Error(err))
	}
	log.Info("Stats Writer consuming", zap.Strings("topics", handler.StatsTopics()), zap.Strings("brokers", brokers))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Stats Writer...")
	if err := uc.Close(); err != nil {
		log.Error("Failed to close stats usecase", zap.Error(err))
	}
}
