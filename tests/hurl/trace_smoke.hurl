# Trace smoke E2E — full chain: merchant → cashier → category → product → order → create transaction
# Jalankan (dari tests/smoke/trace_smoke.sh, bukan manual):
#   hurl --variable baseUrl=http://localhost:5000 --variable uuid=$UUID \
#        --variable traceparent="00-<trace-id>-<span-id>-01" \
#        --test tests/hurl/trace_smoke.hurl
#
# Khusus request create-transaction membawa header `traceparent` dengan
# trace-id tetap. Trace yang dihasilkan harus kontinu di Jaeger:
#   apigateway:post-transaction-create → transaction-service:CreateTransaction
#   → Kafka email-service-topic-transaction-create → email-service:consume:...
# Catatan: amount dihitung ulang service (total + PPN 11%); request amount harus >= itu.

POST {{baseUrl}}/api/auth/register
{
    "firstname": "Trace",
    "lastname": "Smoke",
    "email": "trace.smoke.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "trace.smoke.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Trace {{uuid}}",
    "description": "Toko untuk smoke test trace",
    "address": "Jl. Trace No.1",
    "contact_email": "trace.{{uuid}}@example.com",
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
    "name": "Kasir Trace {{uuid}}"
}
HTTP 201
[Captures]
cashierId: jsonpath "$.data.id"

POST {{baseUrl}}/api/category/create
Authorization: Bearer {{accessToken}}
{
    "name": "Minuman Trace {{uuid}}",
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
name: Es Teh Trace {{uuid}}
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

# Request kunci: membawa traceparent agar trace-id dapat diverifikasi di Jaeger
POST {{baseUrl}}/api/transaction/create
Authorization: Bearer {{accessToken}}
traceparent: {{traceparent}}
{
    "order_id": {{orderId}},
    "cashier_id": {{cashierId}},
    "payment_method": "cash",
    "amount": 30000
}
HTTP 200
[Captures]
transactionId: jsonpath "$.data.id"

GET {{baseUrl}}/api/transaction
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/transaction/yearly-success?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/trashed/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/transaction/restore/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/transaction/permanent/{{transactionId}}
Authorization: Bearer {{accessToken}}
HTTP 200
