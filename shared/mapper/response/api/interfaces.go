package response_api

import (
	authpb "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	cashierpb "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	categorypb "github.com/MamangRust/microservice-point-of-sale-pb/category"
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

type AuthResponseMapper interface {
	ToResponseVerifyCode(res *authpb.ApiResponseVerifyCode) *response.ApiResponseVerifyCode
	ToResponseForgotPassword(res *authpb.ApiResponseForgotPassword) *response.ApiResponseForgotPassword
	ToResponseResetPassword(res *authpb.ApiResponseResetPassword) *response.ApiResponseResetPassword
	ToResponseLogin(res *authpb.ApiResponseLogin) *response.ApiResponseLogin
	ToResponseRegister(res *authpb.ApiResponseRegister) *response.ApiResponseRegister
	ToResponseRefreshToken(res *authpb.ApiResponseRefreshToken) *response.ApiResponseRefreshToken
	ToResponseGetMe(res *authpb.ApiResponseGetMe) *response.ApiResponseGetMe
}

type RoleResponseMapper interface {
	ToApiResponseRoleAll(pbResponse *rolepb.ApiResponseRoleAll) *response.ApiResponseRoleAll
	ToApiResponseRoleDelete(pbResponse *rolepb.ApiResponseRoleDelete) *response.ApiResponseRoleDelete
	ToApiResponseRole(pbResponse *rolepb.ApiResponseRole) *response.ApiResponseRole
	ToApiResponsesRole(pbResponse *rolepb.ApiResponsesRole) *response.ApiResponsesRole
	ToApiResponsePaginationRole(pbResponse *rolepb.ApiResponsePaginationRole) *response.ApiResponsePaginationRole
	ToApiResponsePaginationRoleDeleteAt(pbResponse *rolepb.ApiResponsePaginationRoleDeleteAt) *response.ApiResponsePaginationRoleDeleteAt
}

type UserResponseMapper interface {
	ToApiResponseUserDeleteAt(pbResponse *userpb.ApiResponseUserDeleteAt) *response.ApiResponseUserDeleteAt
	ToApiResponseUser(pbResponse *userpb.ApiResponseUser) *response.ApiResponseUser
	ToApiResponsesUser(pbResponse *userpb.ApiResponsesUser) *response.ApiResponsesUser

	ToApiResponseUserDelete(pbResponse *userpb.ApiResponseUserDelete) *response.ApiResponseUserDelete
	ToApiResponseUserAll(pbResponse *userpb.ApiResponseUserAll) *response.ApiResponseUserAll
	ToApiResponsePaginationUserDeleteAt(pbResponse *userpb.ApiResponsePaginationUserDeleteAt) *response.ApiResponsePaginationUserDeleteAt
	ToApiResponsePaginationUser(pbResponse *userpb.ApiResponsePaginationUser) *response.ApiResponsePaginationUser
}

type CategoryResponseMapper interface {
	ToApiResponseCategory(pbResponse *categorypb.ApiResponseCategory) *response.ApiResponseCategory
	ToApiResponseCategoryDeleteAt(pbResponse *categorypb.ApiResponseCategoryDeleteAt) *response.ApiResponseCategoryDeleteAt
	ToApiResponsesCategory(pbResponse *categorypb.ApiResponsesCategory) *response.ApiResponsesCategory
	ToApiResponseCategoryDelete(pbResponse *categorypb.ApiResponseCategoryDelete) *response.ApiResponseCategoryDelete
	ToApiResponseCategoryAll(pbResponse *categorypb.ApiResponseCategoryAll) *response.ApiResponseCategoryAll
	ToApiResponsePaginationCategoryDeleteAt(pbResponse *categorypb.ApiResponsePaginationCategoryDeleteAt) *response.ApiResponsePaginationCategoryDeleteAt
	ToApiResponsePaginationCategory(pbResponse *categorypb.ApiResponsePaginationCategory) *response.ApiResponsePaginationCategory
}

type CashierResponseMapper interface {
	ToApiResponseCashier(pbResponse *cashierpb.ApiResponseCashier) *response.ApiResponseCashier
	ToApiResponsesCashier(pbResponse *cashierpb.ApiResponsesCashier) *response.ApiResponsesCashier
	ToApiResponseCashierAll(pbResponse *cashierpb.ApiResponseCashierAll) *response.ApiResponseCashierAll
	ToApiResponseCashierDelete(pbResponse *cashierpb.ApiResponseCashierDelete) *response.ApiResponseCashierDelete
	ToApiResponseCashierDeleteAt(pbResponse *cashierpb.ApiResponseCashierDeleteAt) *response.ApiResponseCashierDeleteAt
	ToApiResponsePaginationCashierDeleteAt(pbResponse *cashierpb.ApiResponsePaginationCashierDeleteAt) *response.ApiResponsePaginationCashierDeleteAt
	ToApiResponsePaginationCashier(pbResponse *cashierpb.ApiResponsePaginationCashier) *response.ApiResponsePaginationCashier
}

type MerchantResponseMapper interface {
	ToApiResponseMerchant(pbResponse *merchantpb.ApiResponseMerchant) *response.ApiResponseMerchant

	ToApiResponseMerchantDeleteAt(pbResponse *merchantpb.ApiResponseMerchantDeleteAt) *response.ApiResponseMerchantDeleteAt
	ToApiResponsesMerchant(pbResponse *merchantpb.ApiResponsesMerchant) *response.ApiResponsesMerchant
	ToApiResponseMerchantDelete(pbResponse *merchantpb.ApiResponseMerchantDelete) *response.ApiResponseMerchantDelete
	ToApiResponseMerchantAll(pbResponse *merchantpb.ApiResponseMerchantAll) *response.ApiResponseMerchantAll
	ToApiResponsePaginationMerchantDeleteAt(pbResponse *merchantpb.ApiResponsePaginationMerchantDeleteAt) *response.ApiResponsePaginationMerchantDeleteAt
	ToApiResponsePaginationMerchant(pbResponse *merchantpb.ApiResponsePaginationMerchant) *response.ApiResponsePaginationMerchant
}

type MerchantDocumentResponseMapper interface {
	ToApiResponseMerchantDocument(doc *merchantdocumentpb.ApiResponseMerchantDocument) *response.ApiResponseMerchantDocument
	ToApiResponsesMerchantDocument(docs *merchantdocumentpb.ApiResponsesMerchantDocument) *response.ApiResponsesMerchantDocument

	ToApiResponsePaginationMerchantDocument(docs *merchantdocumentpb.ApiResponsePaginationMerchantDocument) *response.ApiResponsePaginationMerchantDocument
	ToApiResponsePaginationMerchantDocumentDeleteAt(docs *merchantdocumentpb.ApiResponsePaginationMerchantDocumentAt) *response.ApiResponsePaginationMerchantDocumentDeleteAt

	ToApiResponseMerchantDocumentAll(resp *merchantdocumentpb.ApiResponseMerchantDocumentAll) *response.ApiResponseMerchantDocumentAll
	ToApiResponseMerchantDocumentDeleteAt(resp *merchantdocumentpb.ApiResponseMerchantDocumentDelete) *response.ApiResponseMerchantDocumentDelete
}

type OrderItemResponseMapper interface {
	ToApiResponseOrderItem(pbResponse *orderitempb.ApiResponseOrderItem) *response.ApiResponseOrderItem
	ToApiResponsesOrderItem(pbResponse *orderitempb.ApiResponsesOrderItem) *response.ApiResponsesOrderItem
	ToApiResponseOrderItemDelete(pbResponse *orderitempb.ApiResponseOrderItemDelete) *response.ApiResponseOrderItemDelete
	ToApiResponseOrderItemAll(pbResponse *orderitempb.ApiResponseOrderItemAll) *response.ApiResponseOrderItemAll
	ToApiResponsePaginationOrderItemDeleteAt(pbResponse *orderitempb.ApiResponsePaginationOrderItemDeleteAt) *response.ApiResponsePaginationOrderItemDeleteAt
	ToApiResponsePaginationOrderItem(pbResponse *orderitempb.ApiResponsePaginationOrderItem) *response.ApiResponsePaginationOrderItem
}

type OrderResponseMapper interface {
	ToApiResponseOrder(pbResponse *orderpb.ApiResponseOrder) *response.ApiResponseOrder
	ToApiResponseOrderDeleteAt(pbResponse *orderpb.ApiResponseOrderDeleteAt) *response.ApiResponseOrderDeleteAt
	ToApiResponsesOrder(pbResponse *orderpb.ApiResponsesOrder) *response.ApiResponsesOrder
	ToApiResponseOrderDelete(pbResponse *orderpb.ApiResponseOrderDelete) *response.ApiResponseOrderDelete
	ToApiResponseOrderAll(pbResponse *orderpb.ApiResponseOrderAll) *response.ApiResponseOrderAll
	ToApiResponsePaginationOrderDeleteAt(pbResponse *orderpb.ApiResponsePaginationOrderDeleteAt) *response.ApiResponsePaginationOrderDeleteAt
	ToApiResponsePaginationOrder(pbResponse *orderpb.ApiResponsePaginationOrder) *response.ApiResponsePaginationOrder
}

type ProductResponseMapper interface {
	ToApiResponseProduct(pbResponse *productpb.ApiResponseProduct) *response.ApiResponseProduct
	ToApiResponsesProductDeleteAt(pbResponse *productpb.ApiResponseProductDeleteAt) *response.ApiResponseProductDeleteAt
	ToApiResponsesProduct(pbResponse *productpb.ApiResponsesProduct) *response.ApiResponsesProduct
	ToApiResponseProductDelete(pbResponse *productpb.ApiResponseProductDelete) *response.ApiResponseProductDelete
	ToApiResponseProductAll(pbResponse *productpb.ApiResponseProductAll) *response.ApiResponseProductAll
	ToApiResponsePaginationProductDeleteAt(pbResponse *productpb.ApiResponsePaginationProductDeleteAt) *response.ApiResponsePaginationProductDeleteAt
	ToApiResponsePaginationProduct(pbResponse *productpb.ApiResponsePaginationProduct) *response.ApiResponsePaginationProduct
}

type TransactionResponseMapper interface {
	ToApiResponseTransaction(pbResponse *transactionpb.ApiResponseTransaction) *response.ApiResponseTransaction
	ToApiResponseTransactionDeleteAt(pbResponse *transactionpb.ApiResponseTransactionDeleteAt) *response.ApiResponseTransactionDeleteAt
	ToApiResponsesTransaction(pbResponse *transactionpb.ApiResponsesTransaction) *response.ApiResponsesTransaction
	ToApiResponseTransactionDelete(pbResponse *transactionpb.ApiResponseTransactionDelete) *response.ApiResponseTransactionDelete
	ToApiResponseTransactionAll(pbResponse *transactionpb.ApiResponseTransactionAll) *response.ApiResponseTransactionAll
	ToApiResponsePaginationTransactionDeleteAt(pbResponse *transactionpb.ApiResponsePaginationTransactionDeleteAt) *response.ApiResponsePaginationTransactionDeleteAt
	ToApiResponsePaginationTransaction(pbResponse *transactionpb.ApiResponsePaginationTransaction) *response.ApiResponsePaginationTransaction
}
