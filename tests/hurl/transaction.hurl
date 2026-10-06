# Transaction E2E — full chain: merchant → cashier → category → product → order → create transaction
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/transaction.hurl
# Catatan: amount dihitung ulang service (total + PPN 11%); request amount harus >= itu.

POST {{baseUrl}}/api/auth/register
{
    "firstname": "Transaction",
    "lastname": "Owner",
    "email": "tx.owner.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "tx.owner.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Tx {{uuid}}",
    "description": "Toko untuk test transaksi",
    "address": "Jl. Tx No.5",
    "contact_email": "tx.{{uuid}}@example.com",
    "contact_phone": "081298765429",
    "status": "active"
}
HTTP 200
[Captures]
merchantId: jsonpath "$.data.id"

POST {{baseUrl}}/api/cashier/create
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "user_id": {{ownerId}},
    "name": "Kasir Tx {{uuid}}"
}
HTTP 201
[Captures]
cashierId: jsonpath "$.data.id"

POST {{baseUrl}}/api/category/create
Authorization: Bearer {{accessToken}}
{
    "name": "Minuman Tx {{uuid}}",
    "description": "Kategori minuman"
}
HTTP 201
[Captures]
categoryId: jsonpath "$.data.id"

POST {{baseUrl}}/api/product/create
Authorization: Bearer {{accessToken}}
[MultipartFormData]
merchant_id: {{merchantId}}
category_id: {{categoryId}}
name: Es Teh {{uuid}}
description: Es teh manis
price: 10000
count_in_stock: 100
brand: Tehku
weight: 200
image_product: file, fixtures/product.jpg; image/jpeg
HTTP 201
[Captures]
productId: jsonpath "$.data.id"

# Order: total 20000 + PPN 2200 = 22200 → amount 30000 cukup
POST {{baseUrl}}/api/order/create
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "cashier_id": {{cashierId}},
    "items": [
        {
            "product_id": {{productId}},
            "quantity": 2
        }
    ]
}
HTTP 200
[Captures]
orderId: jsonpath "$.data.id"

POST {{baseUrl}}/api/transaction/create
Authorization: Bearer {{accessToken}}
{
    "order_id": {{orderId}},
    "cashier_id": {{cashierId}},
    "payment_method": "cash",
    "amount": 30000
}
HTTP 200
[Captures]
transactionId: jsonpath "$.data.id"

# Update transaksi — payment_status hasil create = "success" (bukan paid/refunded) → boleh update
POST {{baseUrl}}/api/transaction/update/{{transactionId}}
Authorization: Bearer {{accessToken}}
{
    "order_id": {{orderId}},
    "cashier_id": {{cashierId}},
    "payment_method": "cash",
    "amount": 30000
}
HTTP 200

GET {{baseUrl}}/api/transaction
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: status ──
GET {{baseUrl}}/api/transaction/monthly-success?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/yearly-success?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/monthly-failed?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/yearly-failed?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/monthly-success?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/yearly-success?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/monthly-failed?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/yearly-failed?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: method ──
GET {{baseUrl}}/api/transaction/monthly-method-success?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/yearly-method-success?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/monthly-method-failed?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/yearly-method-failed?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/monthly-method-success/{{merchantId}}?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/yearly-method-success/{{merchantId}}?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/monthly-method-failed/{{merchantId}}?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/merchant/yearly-method-failed/{{merchantId}}?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/trashed/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/restore/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Trash lagi → restore/all → trash lagi → permanent/all
POST {{baseUrl}}/api/transaction/trashed/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/trashed/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200

# Transaksi kedua — coverage DELETE /permanent/:id (sibling dari permanent/all).
# PENTING: tetap setelah permanent/all — ikuti pola yang sama seperti order.hurl.
POST {{baseUrl}}/api/transaction/create
Authorization: Bearer {{accessToken}}
{
    "order_id": {{orderId}},
    "cashier_id": {{cashierId}},
    "payment_method": "cash",
    "amount": 30000
}
HTTP 200
[Captures]
transactionId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/transaction/trashed/{{transactionId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/transaction/permanent/{{transactionId2}}
Authorization: Bearer {{accessToken}}
HTTP 200
