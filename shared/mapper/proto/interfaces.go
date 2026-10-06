package protomapper

import (
	authpb "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
	merchantpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	merchantdocumentpb "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	orderpb "github.com/MamangRust/microservice-point-of-sale-pb/order"
	orderitempb "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	productpb "github.com/MamangRust/microservice-point-of-sale-pb/product"
	rolepb "github.com/MamangRust/microservice-point-of-sale-pb/role"
	transactionpb "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	userpb "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type AuthProtoMapper interface {
	ToProtoResponseVerifyCode(status string, message string) *authpb.ApiResponseVerifyCode
	ToProtoResponseForgotPassword(status string, message string) *authpb.ApiResponseForgotPassword
	ToProtoResponseResetPassword(status string, message string) *authpb.ApiResponseResetPassword
	ToProtoResponseLogin(status string, message string, response *response.TokenResponse) *authpb.ApiResponseLogin
	ToProtoResponseRegister(status string, message string, response *response.UserResponse) *authpb.ApiResponseRegister
	ToProtoResponseRefreshToken(status string, message string, response *response.TokenResponse) *authpb.ApiResponseRefreshToken
	ToProtoResponseGetMe(status string, message string, response *response.UserResponse) *authpb.ApiResponseGetMe
}

type UserProtoMapper interface {
	ToProtoResponseUserDeleteAt(status string, message string, pbResponse *response.UserResponseDeleteAt) *userpb.ApiResponseUserDeleteAt
	ToProtoResponsesUser(status string, message string, pbResponse []*response.UserResponse) *userpb.ApiResponsesUser
	ToProtoResponseUser(status string, message string, pbResponse *response.UserResponse) *userpb.ApiResponseUser
	ToProtoResponseUserDelete(status string, message string) *userpb.ApiResponseUserDelete
	ToProtoResponseUserAll(status string, message string) *userpb.ApiResponseUserAll
	ToProtoResponsePaginationUserDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, users []*response.UserResponseDeleteAt) *userpb.ApiResponsePaginationUserDeleteAt
	ToProtoResponsePaginationUser(pagination *commonpb.PaginationMeta, status string, message string, users []*response.UserResponse) *userpb.ApiResponsePaginationUser
}

type RoleProtoMapper interface {
	ToProtoResponseRoleAll(status string, message string) *rolepb.ApiResponseRoleAll
	ToProtoResponseRoleDelete(status string, message string) *rolepb.ApiResponseRoleDelete
	ToProtoResponseRole(status string, message string, pbResponse *response.RoleResponse) *rolepb.ApiResponseRole
	ToProtoResponsesRole(status string, message string, pbResponse []*response.RoleResponse) *rolepb.ApiResponsesRole
	ToProtoResponsePaginationRole(pagination *commonpb.PaginationMeta, status string, message string, pbResponse []*response.RoleResponse) *rolepb.ApiResponsePaginationRole
	ToProtoResponsePaginationRoleDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, pbResponse []*response.RoleResponseDeleteAt) *rolepb.ApiResponsePaginationRoleDeleteAt
}

type CategoryProtoMapper interface {
	ToProtoResponsesCategory(status string, message string, pbResponse []*response.CategoryResponse) *categorypb.ApiResponsesCategory
	ToProtoResponseCategoryDeleteAt(status string, message string, pbResponse *response.CategoryResponseDeleteAt) *categorypb.ApiResponseCategoryDeleteAt

	ToProtoResponseCategoryAll(status string, message string) *categorypb.ApiResponseCategoryAll
	ToProtoResponseCategory(status string, message string, pbResponse *response.CategoryResponse) *categorypb.ApiResponseCategory
	ToProtoResponseCategoryDelete(status string, message string) *categorypb.ApiResponseCategoryDelete
	ToProtoResponsePaginationCategoryDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, categories []*response.CategoryResponseDeleteAt) *categorypb.ApiResponsePaginationCategoryDeleteAt
	ToProtoResponsePaginationCategory(pagination *commonpb.PaginationMeta, status string, message string, categories []*response.CategoryResponse) *categorypb.ApiResponsePaginationCategory
}

type CashierProtoMapper interface {
	ToProtoResponseCashier(status string, message string, pbResponse *response.CashierResponse) *cashierpb.ApiResponseCashier
	ToProtoResponseCashierDeleteAt(status string, message string, pbResponse *response.CashierResponseDeleteAt) *cashierpb.ApiResponseCashierDeleteAt
	ToProtoResponsesCashier(status string, message string, pbResponse []*response.CashierResponse) *cashierpb.ApiResponsesCashier
	ToProtoResponseCashierDelete(status string, message string) *cashierpb.ApiResponseCashierDelete
	ToProtoResponseCashierAll(status string, message string) *cashierpb.ApiResponseCashierAll
	ToProtoResponsePaginationCashierDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, users []*response.CashierResponseDeleteAt) *cashierpb.ApiResponsePaginationCashierDeleteAt
	ToProtoResponsePaginationCashier(pagination *commonpb.PaginationMeta, status string, message string, users []*response.CashierResponse) *cashierpb.ApiResponsePaginationCashier
}

type MerchantProtoMapper interface {
	ToProtoResponseMerchant(status string, message string, pbResponse *response.MerchantResponse) *merchantpb.ApiResponseMerchant
	ToProtoResponseMerchantDeleteAt(status string, message string, pbResponse *response.MerchantResponseDeleteAt) *merchantpb.ApiResponseMerchantDeleteAt

	ToProtoResponsesMerchant(status string, message string, pbResponse []*response.MerchantResponse) *merchantpb.ApiResponsesMerchant
	ToProtoResponseMerchantDelete(status string, message string) *merchantpb.ApiResponseMerchantDelete
	ToProtoResponseMerchantAll(status string, message string) *merchantpb.ApiResponseMerchantAll
	ToProtoResponsePaginationMerchantDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, merchants []*response.MerchantResponseDeleteAt) *merchantpb.ApiResponsePaginationMerchantDeleteAt
	ToProtoResponsePaginationMerchant(pagination *commonpb.PaginationMeta, status string, message string, merchants []*response.MerchantResponse) *merchantpb.ApiResponsePaginationMerchant
}

type MerchantDocumentProtoMapper interface {
	ToProtoResponseMerchantDocument(status string, message string, doc *response.MerchantDocumentResponse) *merchantdocumentpb.ApiResponseMerchantDocument
	ToProtoResponsesMerchantDocument(status string, message string, docs []*response.MerchantDocumentResponse) *merchantdocumentpb.ApiResponsesMerchantDocument

	ToProtoResponsePaginationMerchantDocument(pagination *commonpb.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponse) *merchantdocumentpb.ApiResponsePaginationMerchantDocument
	ToProtoResponsePaginationMerchantDocumentDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponseDeleteAt) *merchantdocumentpb.ApiResponsePaginationMerchantDocumentAt

	ToProtoResponseMerchantDocumentDelete(status string, message string) *merchantdocumentpb.ApiResponseMerchantDocumentDelete

	ToProtoResponseMerchantDocumentAll(status string, message string) *merchantdocumentpb.ApiResponseMerchantDocumentAll
}

type OrderItemProtoMapper interface {
	ToProtoResponseOrderItem(status string, message string, pbResponse *response.OrderItemResponse) *orderitempb.ApiResponseOrderItem
	ToProtoResponsesOrderItem(status string, message string, pbResponse []*response.OrderItemResponse) *orderitempb.ApiResponsesOrderItem
	ToProtoResponseOrderItemDelete(status string, message string) *orderitempb.ApiResponseOrderItemDelete
	ToProtoResponseOrderItemAll(status string, message string) *orderitempb.ApiResponseOrderItemAll
	ToProtoResponsePaginationOrderItemDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, orderItems []*response.OrderItemResponseDeleteAt) *orderitempb.ApiResponsePaginationOrderItemDeleteAt
	ToProtoResponsePaginationOrderItem(pagination *commonpb.PaginationMeta, status string, message string, orderItems []*response.OrderItemResponse) *orderitempb.ApiResponsePaginationOrderItem
}

type OrderProtoMapper interface {
	ToProtoResponseOrder(status string, message string, pbResponse *response.OrderResponse) *orderpb.ApiResponseOrder
	ToProtoResponseOrderDeleteAt(status string, message string, pbResponse *response.OrderResponseDeleteAt) *orderpb.ApiResponseOrderDeleteAt
	ToProtoResponsesOrder(status string, message string, pbResponse []*response.OrderResponse) *orderpb.ApiResponsesOrder
	ToProtoResponseOrderDelete(status string, message string) *orderpb.ApiResponseOrderDelete
	ToProtoResponseOrderAll(status string, message string) *orderpb.ApiResponseOrderAll
	ToProtoResponsePaginationOrderDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, orders []*response.OrderResponseDeleteAt) *orderpb.ApiResponsePaginationOrderDeleteAt
	ToProtoResponsePaginationOrder(pagination *commonpb.PaginationMeta, status string, message string, orders []*response.OrderResponse) *orderpb.ApiResponsePaginationOrder
}

type ProductProtoMapper interface {
	ToProtoResponseProduct(status string, message string, pbResponse *response.ProductResponse) *productpb.ApiResponseProduct
	ToProtoResponseProductDeleteAt(status string, message string, pbResponse *response.ProductResponseDeleteAt) *productpb.ApiResponseProductDeleteAt

	ToProtoResponsesProduct(status string, message string, pbResponse []*response.ProductResponse) *productpb.ApiResponsesProduct
	ToProtoResponseProductDelete(status string, message string) *productpb.ApiResponseProductDelete
	ToProtoResponseProductAll(status string, message string) *productpb.ApiResponseProductAll
	ToProtoResponsePaginationProductDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, products []*response.ProductResponseDeleteAt) *productpb.ApiResponsePaginationProductDeleteAt
	ToProtoResponsePaginationProduct(pagination *commonpb.PaginationMeta, status string, message string, products []*response.ProductResponse) *productpb.ApiResponsePaginationProduct
}

type TransactionProtoMapper interface {
	ToProtoResponseTransaction(status string, message string, trans *response.TransactionResponse) *transactionpb.ApiResponseTransaction
	ToProtoResponseTransactionDeleteAt(status string, message string, trans *response.TransactionResponseDeleteAt) *transactionpb.ApiResponseTransactionDeleteAt
	ToProtoResponsesTransaction(status string, message string, transList []*response.TransactionResponse) *transactionpb.ApiResponsesTransaction
	ToProtoResponseTransactionDelete(status string, message string) *transactionpb.ApiResponseTransactionDelete
	ToProtoResponseTransactionAll(status string, message string) *transactionpb.ApiResponseTransactionAll
	ToProtoResponsePaginationTransactionDeleteAt(pagination *commonpb.PaginationMeta, status string, message string, transactions []*response.TransactionResponseDeleteAt) *transactionpb.ApiResponsePaginationTransactionDeleteAt
	ToProtoResponsePaginationTransaction(pagination *commonpb.PaginationMeta, status string, message string, transactions []*response.TransactionResponse) *transactionpb.ApiResponsePaginationTransaction
}
