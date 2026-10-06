package apps

import (
	"context"
	"os"
	"time"

	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	mencache "github.com/MamangRust/microservice-point-of-sale-product/cache"
	"github.com/MamangRust/microservice-point-of-sale-product/handler"
	"github.com/MamangRust/microservice-point-of-sale-product/repository"
	"github.com/MamangRust/microservice-point-of-sale-product/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	categoryAddr := getEnv("GRPC_CATEGORY_ADDR", "localhost:50054")
	merchantAddr := getEnv("GRPC_MERCHANT_ADDR", "localhost:50056")

	categoryConn, err := server.NewGRPCClient(categoryAddr)
	if err != nil {
		return nil, err
	}

	merchantConn, err := server.NewGRPCClient(merchantAddr)
	if err != nil {
		categoryConn.Close()
		return nil, err
	}

	go func() {
		<-srv.Ctx.Done()
		srv.Logger.Info("Closing gRPC client connections in Product microservice")
		categoryConn.Close()
		merchantConn.Close()
	}()

	categoryClient := pbcategory.NewCategoryServiceClient(categoryConn)
	merchantClient := pbmerchant.NewMerchantServiceClient(merchantConn)

	guardCategory := resilience.NewDependencyGuard("category", 5, 30, 100, 3*time.Second, srv.Logger)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(
		srv.MongoDB, categoryClient, merchantClient,
		repository.GuardOptions{
			Category: []adapter.GuardOption{adapter.WithDependencyGuard(guardCategory)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)

	mencacheObj := mencache.NewMencache(srv.CacheStore)

	obs := observability.NewTraceLoggerObservability(srv.Logger)

	services := service.NewService(&service.Deps{
		Mencache:      mencacheObj,
		Ctx:           context.Background(),
		Repositories:  repos,
		Logger:        srv.Logger,
		Observability: obs,
	})

	handlers := handler.NewHandler(services)

	srv.RegisterServices = func(gs *grpc.Server) {
		pbproduct.RegisterProductServiceServer(gs, handlers)
	}

	srv.Logger.Info("Connected to gRPC microservices from Product microservice",
		zap.String("category", categoryAddr),
		zap.String("merchant", merchantAddr),
	)

	return srv, nil
}
