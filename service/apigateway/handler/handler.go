package handler

import (
	"fmt"
	"strconv"

	auth_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/auth"
	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	authpb "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	merchantpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	merchantdocumentpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	orderpb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	orderitempb "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	productpb "github.com/MamangRust/microservice-point-of-sale-pb/product"
	rolepb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	userrolepb "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	statspb "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	userpb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/upload_image"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"

	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
)

type ServiceConnections struct {
	Auth        *grpc.ClientConn
	Role        *grpc.ClientConn
	User        *grpc.ClientConn
	Cashier     *grpc.ClientConn
	Category    *grpc.ClientConn
	Merchant    *grpc.ClientConn
	OrderItem   *grpc.ClientConn
	Order       *grpc.ClientConn
	Product     *grpc.ClientConn
	Transaction *grpc.ClientConn
	StatsReader *grpc.ClientConn
}

type Deps struct {
	Token              auth.TokenManager
	E                  *echo.Echo
	Logger             logger.LoggerInterface
	Mapping            *response_api.ResponseApiMapper
	ImageUpload        upload_image.ImageUploads
	ServiceConnections *ServiceConnections
	ApiHandler         errors.ApiHandler
	AuthCache          auth_cache.AuthMencache
	GatewayCache       *gateway_cache.GatewayCache
}

func NewHandler(deps *Deps) {
	clientAuth := authpb.NewAuthServiceClient(deps.ServiceConnections.Auth)
	clientRoleQuery := rolepb.NewRoleServiceClient(deps.ServiceConnections.Role)
	clientRoleCommand := rolepb.NewRoleServiceClient(deps.ServiceConnections.Role)
	clientUserRole := userrolepb.NewUserRoleServiceClient(deps.ServiceConnections.Role)
	clientUserQuery := userpb.NewUserServiceClient(deps.ServiceConnections.User)
	clientUserCommand := userpb.NewUserServiceClient(deps.ServiceConnections.User)
	clientCategoryQuery := categorypb.NewCategoryServiceClient(deps.ServiceConnections.Category)
	clientCategoryCommand := categorypb.NewCategoryServiceClient(deps.ServiceConnections.Category)
	clientCashierQuery := cashierpb.NewCashierServiceClient(deps.ServiceConnections.Cashier)
	clientCashierCommand := cashierpb.NewCashierServiceClient(deps.ServiceConnections.Cashier)
	clientMerchantQuery := merchantpb.NewMerchantServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantCommand := merchantpb.NewMerchantServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantDocumentQuery := merchantdocumentpb.NewMerchantDocumentServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantDocumentCommand := merchantdocumentpb.NewMerchantDocumentServiceClient(deps.ServiceConnections.Merchant)
	clientOrderItemQuery := orderitempb.NewOrderItemServiceClient(deps.ServiceConnections.OrderItem)
	clientOrderQuery := orderpb.NewOrderServiceClient(deps.ServiceConnections.Order)
	clientOrderCommand := orderpb.NewOrderServiceClient(deps.ServiceConnections.Order)
	clientProductQuery := productpb.NewProductServiceClient(deps.ServiceConnections.Product)
	clientProductCommand := productpb.NewProductServiceClient(deps.ServiceConnections.Product)
	clientTransactionQuery := transactionpb.NewTransactionServiceClient(deps.ServiceConnections.Transaction)
	clientTransactionCommand := transactionpb.NewTransactionServiceClient(deps.ServiceConnections.Transaction)

	NewHandlerAuth(deps.E, clientAuth, deps.Logger, deps.Mapping.AuthResponseMapper, deps.ApiHandler, deps.AuthCache)
	NewHandlerRole(deps.E, clientRoleQuery, clientRoleCommand, clientUserRole, deps.Logger, deps.Mapping.RoleResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerUser(deps.E, clientUserQuery, clientUserCommand, deps.Logger, deps.Mapping.UserResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCategory(deps.E, clientCategoryQuery, clientCategoryCommand, deps.Logger, deps.Mapping.CategoryResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCashier(deps.E, clientCashierQuery, clientCashierCommand, deps.Logger, deps.Mapping.CashierResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchant(deps.E, clientMerchantQuery, clientMerchantCommand, deps.Logger, deps.Mapping.MerchantResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchantDocument(deps.E, clientMerchantDocumentQuery, clientMerchantDocumentCommand, deps.Logger, deps.Mapping.MerchantDocumentProMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrderItem(deps.E, clientOrderItemQuery, deps.Logger, deps.Mapping.OrderItemResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrder(deps.E, clientOrderQuery, clientOrderCommand, deps.Logger, deps.Mapping.OrderResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerProduct(deps.E, clientProductQuery, clientProductCommand, deps.Logger, deps.Mapping.ProductResponseMapper, deps.ImageUpload, deps.ApiHandler, deps.GatewayCache)
	NewHandlerTransaction(deps.E, clientTransactionQuery, clientTransactionCommand, deps.Logger, deps.Mapping.TransactionResponseMapper, deps.ApiHandler, deps.GatewayCache)

	statsClients := &StatsClients{
		Order:              statspb.NewOrderStatsServiceClient(deps.ServiceConnections.StatsReader),
		OrderByMerchant:    statspb.NewOrderStatsByMerchantServiceClient(deps.ServiceConnections.StatsReader),
		OrderById:          statspb.NewOrderStatsByIdServiceClient(deps.ServiceConnections.StatsReader),
		Cashier:            statspb.NewCashierStatsServiceClient(deps.ServiceConnections.StatsReader),
		CashierByMerchant:  statspb.NewCashierStatsByMerchantServiceClient(deps.ServiceConnections.StatsReader),
		CashierById:        statspb.NewCashierStatsByIdServiceClient(deps.ServiceConnections.StatsReader),
		Category:           statspb.NewCategoryStatsServiceClient(deps.ServiceConnections.StatsReader),
		CategoryByMerchant: statspb.NewCategoryStatsByMerchantServiceClient(deps.ServiceConnections.StatsReader),
		CategoryById:       statspb.NewCategoryStatsByIdServiceClient(deps.ServiceConnections.StatsReader),
		TransactionStatus:  statspb.NewTransactionStatsStatusServiceClient(deps.ServiceConnections.StatsReader),
		TransactionMethod:  statspb.NewTransactionStatsMethodServiceClient(deps.ServiceConnections.StatsReader),
	}

	NewHandlerStats(deps.E, statsClients, deps.Logger, deps.Mapping.StatsResponseMapper, deps.ApiHandler, deps.GatewayCache)
}

func parseQueryInt(c echo.Context, key string, defaultValue int) int {
	val, err := strconv.Atoi(c.QueryParam(key))
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func parseQueryIntWithValidation(c echo.Context, key string, min, max int) (int, error) {
	valStr := c.QueryParam(key)
	val, err := strconv.Atoi(valStr)
	if err != nil || val < min || val > max {
		return 0, fmt.Errorf("invalid %s: %s", key, valStr)
	}
	return val, nil
}
