package apps

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	auth_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/auth"
	"github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	"github.com/MamangRust/microservice-point-of-sale-apigateway/middlewares"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	otel_pkg "github.com/MamangRust/microservice-point-of-sale-pkg/otel"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	"github.com/MamangRust/microservice-point-of-sale-pkg/upload_image"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	shared_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServiceAddresses struct {
	Auth        string
	Role        string
	User        string
	Category    string
	Cashier     string
	Merchant    string
	OrderItem   string
	Order       string
	Product     string
	Transaction string
	StatsReader string
}

func loadServiceAddresses() ServiceAddresses {
	return ServiceAddresses{
		Auth:        getEnvOrDefault("GRPC_AUTH_ADDR", "localhost:50051"),
		Role:        getEnvOrDefault("GRPC_ROLE_ADDR", "localhost:50052"),
		User:        getEnvOrDefault("GRPC_USER_ADDR", "localhost:50053"),
		Cashier:     getEnvOrDefault("GRPC_CASHIER_ADDR", "localhost:50055"),
		Category:    getEnvOrDefault("GRPC_CATEGORY_ADDR", "localhost:50054"),
		Merchant:    getEnvOrDefault("GRPC_MERCHANT_ADDR", "localhost:50056"),
		OrderItem:   getEnvOrDefault("GRPC_ORDERITEM_ADDR", "localhost:50057"),
		Order:       getEnvOrDefault("GRPC_ORDER_ADDR", "localhost:50058"),
		Product:     getEnvOrDefault("GRPC_PRODUCT_ADDR", "localhost:50059"),
		Transaction: getEnvOrDefault("GRPC_TRANSACTION_ADDR", "localhost:50060"),
		StatsReader: getEnvOrDefault("GRPC_STATS_READER_ADDR", "localhost:50070"),
	}
}

func createServiceConnections(addresses ServiceAddresses, logger logger.LoggerInterface) (handler.ServiceConnections, error) {
	var connections handler.ServiceConnections

	conns := map[string]*string{
		"Auth":        &addresses.Auth,
		"Role":        &addresses.Role,
		"User":        &addresses.User,
		"Cashier":     &addresses.Cashier,
		"Category":    &addresses.Category,
		"Merchant":    &addresses.Merchant,
		"OrderItem":   &addresses.OrderItem,
		"Order":       &addresses.Order,
		"Product":     &addresses.Product,
		"Transaction": &addresses.Transaction,
		"StatsReader": &addresses.StatsReader,
	}

	for name, addr := range conns {
		conn, err := createConnection(*addr, name, logger)
		if err != nil {
			return connections, err
		}
		switch name {
		case "Auth":
			connections.Auth = conn
		case "Role":
			connections.Role = conn
		case "User":
			connections.User = conn
		case "Cashier":
			connections.Cashier = conn
		case "Category":
			connections.Category = conn
		case "Merchant":
			connections.Merchant = conn
		case "OrderItem":
			connections.OrderItem = conn
		case "Order":
			connections.Order = conn
		case "Product":
			connections.Product = conn
		case "Transaction":
			connections.Transaction = conn
		case "StatsReader":
			connections.StatsReader = conn
		}
	}

	return connections, nil
}

// @title PointOfsale gRPC
// @version 1.0
// @description gRPC based Point Of Sale service

// @host localhost:5000
// @BasePath /api/

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token obtained from login
// @security ApiKeyAuth
type Client struct {
	App    *echo.Echo
	Logger logger.LoggerInterface
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.App.Shutdown(ctx)
}

func RunClient() (*Client, func(), error) {
	flag.Parse()

	if err := dotenv.Viper(); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	endpoint := viper.GetString("OTEL_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4317"
	}

	// Telemetry must be initialized before the logger: the logger attaches an
	// OTLP log bridge to the LoggerProvider (otelzap).
	telemetry := otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:            "apigateway",
		ServiceVersion:         "1.0.0",
		Environment:            viper.GetString("ENV"),
		Endpoint:               endpoint,
		Insecure:               true,
		EnableRuntimeMetrics:   true,
		RuntimeMetricsInterval: 15 * time.Second,
		Disabled:               os.Getenv("OTEL_ENABLED") == "false",
	})
	if err := telemetry.Init(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	log, err := logger.NewLogger("apigateway", telemetry.GetLogger())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	addresses := loadServiceAddresses()

	log.Debug("Creating gRPC connections...")
	conns, err := createServiceConnections(addresses, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect services: %w", err)
	}

	e := setupEcho()

	token, err := auth.NewManager(viper.GetString("SECRET_KEY"))
	if err != nil {
		log.Fatal("Failed to create token manager", zap.Error(err))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT")),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	})

	cacheMetrics, err := observability.NewCacheMetrics("apigateway")
	if err != nil {
		log.Fatal("Failed to initialize cache metrics", zap.Error(err))
	}

	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)
	authMencache := auth_cache.NewMencache(cacheStore)
	gatewayCache := gateway_cache.NewGatewayCache(cacheStore)

	mapping := response_api.NewResponseApiMapper()
	image_upload := upload_image.NewImageUpload(log)

	obs := observability.NewTraceLoggerObservability(log)
	apiHandler := shared_errors.NewApiHandler(obs, log)

	depsHandler := &handler.Deps{
		Token:              token,
		E:                  e,
		Logger:             log,
		Mapping:            mapping,
		ImageUpload:        image_upload,
		ServiceConnections: &conns,
		ApiHandler:         apiHandler,
		AuthCache:          authMencache,
		GatewayCache:       gatewayCache,
	}

	handler.NewHandler(depsHandler)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Info("Starting API Gateway server on :5000")
		if err := e.Start(":5000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Echo server error", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		log.Info("Starting Prometheus metrics server on :8091")
		if err := http.ListenAndServe(":8091", promhttp.Handler()); err != nil {
			log.Fatal("Metrics server error", zap.Error(err))
		}
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("Shutting down API Gateway...")
		if err := e.Shutdown(ctx); err != nil {
			log.Error("Echo shutdown failed", zap.Error(err))
		}

		closeConnections(conns, log)

		if err := telemetry.Shutdown(context.Background()); err != nil {
			log.Error("Telemetry shutdown failed", zap.Error(err))
		}
	}

	return &Client{App: e, Logger: log}, shutdown, nil
}

func setupEcho() *echo.Echo {
	e := echo.New()

	limiter := middlewares.NewRateLimiter(20, 200)
	e.Use(limiter.Limit, middleware.Recover(), middleware.Logger())

	// Extract incoming W3C trace context (traceparent/tracestate) from HTTP
	// headers so the trace continues from NGINX/client through the gateway.
	e.Use(middlewares.TraceContextMiddleware())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:1420", "http://localhost:33451"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key"},
		AllowCredentials: true,
	}))

	middlewares.WebSecurityConfig(e)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

func createConnection(address, serviceName string, logger logger.LoggerInterface) (*grpc.ClientConn, error) {
	logger.Info(fmt.Sprintf("Connecting to %s service at %s", serviceName, address))
	conn, err := server.NewGRPCClient(address)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to %s service", serviceName), zap.Error(err))
		return nil, err
	}
	return conn, nil
}

func closeConnections(conns handler.ServiceConnections, log logger.LoggerInterface) {
	connsMap := map[string]*grpc.ClientConn{
		"Auth":        conns.Auth,
		"Role":        conns.Role,
		"User":        conns.User,
		"Cashier":     conns.Cashier,
		"Category":    conns.Category,
		"Merchant":    conns.Merchant,
		"OrderItem":   conns.OrderItem,
		"Order":       conns.Order,
		"Product":     conns.Product,
		"Transaction": conns.Transaction,
		"StatsReader": conns.StatsReader,
	}

	for name, conn := range connsMap {
		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Error(fmt.Sprintf("Failed to close %s connection", name), zap.Error(err))
			}
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
