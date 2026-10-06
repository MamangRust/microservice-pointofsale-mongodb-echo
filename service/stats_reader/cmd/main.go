package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	"github.com/MamangRust/microservice-point-of-sale-pkg/clickhouse"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/MamangRust/microservice-point-of-sale-stats-reader/handler"
	"github.com/MamangRust/microservice-point-of-sale-stats-reader/repository"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		zap.L().Error("Failed to load configuration", zap.Error(err))
	}
	log, _ := logger.NewLogger("stats-reader", nil)

	// The ClickHouse database must exist before NewClient can ping.
	if err := clickhouse.EnsureDatabase(log); err != nil {
		log.Fatal("Failed to ensure ClickHouse database", zap.Error(err))
	}
	chConn, err := clickhouse.NewClient(log)
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}
	// Guarantee the stats tables exist even if stats-writer has not run yet.
	if err := clickhouse.ApplySchema(context.Background(), chConn, log); err != nil {
		log.Fatal("Failed to apply ClickHouse schema", zap.Error(err))
	}

	repo := repository.NewClickHouseReaderRepository(chConn)

	// Redis cache (5-minute TTL at the reader layer, §7.4) is best-effort: if
	// Redis is unreachable the reader keeps serving straight from ClickHouse.
	var readerCache *handler.StatsCache
	if store, ok := newCacheStore(log); ok {
		readerCache = handler.NewStatsCache(store)
	}

	orderStatsHandler := handler.NewOrderStatsHandler(repo, readerCache, log)
	cashierStatsHandler := handler.NewCashierStatsHandler(repo, readerCache, log)
	categoryStatsHandler := handler.NewCategoryStatsHandler(repo, readerCache, log)
	transactionStatsHandler := handler.NewTransactionStatsHandler(repo, readerCache, log)

	grpcServer := grpc.NewServer()

	statspb.RegisterOrderStatsServiceServer(grpcServer, orderStatsHandler)
	statspb.RegisterOrderStatsByMerchantServiceServer(grpcServer, orderStatsHandler)
	statspb.RegisterOrderStatsByIdServiceServer(grpcServer, orderStatsHandler)
	statspb.RegisterCashierStatsServiceServer(grpcServer, cashierStatsHandler)
	statspb.RegisterCashierStatsByMerchantServiceServer(grpcServer, cashierStatsHandler)
	statspb.RegisterCashierStatsByIdServiceServer(grpcServer, cashierStatsHandler)
	statspb.RegisterCategoryStatsServiceServer(grpcServer, categoryStatsHandler)
	statspb.RegisterCategoryStatsByMerchantServiceServer(grpcServer, categoryStatsHandler)
	statspb.RegisterCategoryStatsByIdServiceServer(grpcServer, categoryStatsHandler)
	statspb.RegisterTransactionStatsStatusServiceServer(grpcServer, transactionStatsHandler)
	statspb.RegisterTransactionStatsMethodServiceServer(grpcServer, transactionStatsHandler)

	reflection.Register(grpcServer)

	// The listen address is distinct from GRPC_STATS_READER_ADDR: the reader
	// listens on a bind address (":50070") while apigateway dials a routable
	// host:port (e.g. "stats-reader.point-of-sale.svc.cluster.local:50070").
	// STATS_READER_LISTEN_ADDR takes precedence; the dial key is only a
	// convenience fallback for local development (localhost:50070).
	addr := viper.GetString("STATS_READER_LISTEN_ADDR")
	if addr == "" {
		addr = viper.GetString("GRPC_STATS_READER_ADDR")
	}
	if addr == "" {
		addr = ":50070"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err), zap.String("addr", addr))
	}

	log.Info("Stats Reader starting", zap.String("addr", addr))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Stats Reader...")
	grpcServer.GracefulStop()
}

func newCacheStore(log logger.LoggerInterface) (*cache.CacheStore, bool) {
	cacheMetrics, err := observability.NewCacheMetrics("stats-reader")
	if err != nil {
		log.Warn("Cache metrics unavailable; running without Redis cache", zap.Error(err))
		return nil, false
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT")),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn("Redis unavailable; running without Redis cache", zap.Error(err))
		_ = redisClient.Close()
		return nil, false
	}

	return cache.NewCacheStore(redisClient, log, cacheMetrics), true
}
