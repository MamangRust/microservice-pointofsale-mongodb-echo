# Order E2E — full chain: merchant → cashier → category → product → create order → lifecycle
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/order.hurl

POST {{baseUrl}}/api/auth/register
{
    "firstname": "Order",
    "lastname": "Owner",
    "email": "order.owner.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "order.owner.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Order {{uuid}}",
    "description": "Toko untuk test order",
    "address": "Jl. Order No.4",
    "contact_email": "order.{{uuid}}@example.com",
    "contact_phone": "081298765430",
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
    "name": "Kasir Order {{uuid}}"
}
HTTP 201
[Captures]
cashierId: jsonpath "$.data.id"

POST {{baseUrl}}/api/category/create
Authorization: Bearer {{accessToken}}
{
    "name": "Makanan {{uuid}}",
    "description": "Kategori makanan"
}
HTTP 201
[Captures]
categoryId: jsonpath "$.data.id"

POST {{baseUrl}}/api/product/create
Authorization: Bearer {{accessToken}}
[MultipartFormData]
merchant_id: {{merchantId}}
category_id: {{categoryId}}
name: Nasi Goreng {{uuid}}
description: Nasi goreng spesial
price: 30000
count_in_stock: 50
brand: Dapur
weight: 300
image_product: file, fixtures/product.jpg; image/jpeg
HTTP 201
[Captures]
productId: jsonpath "$.data.id"

# Create order — total dihitung service: 2 × qty 2 × 30000 = 120000
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

# Update order — butuh order_item_id (diambil dari order-item milik order ini).
# Catatan: order id fresh per-run (uuid), jadi cache 5-menit gateway tidak pernah basi.
GET {{baseUrl}}/api/order-item/{{orderId}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Captures]
orderItemId: jsonpath "$.data[0].id"

POST {{baseUrl}}/api/order/update/{{orderId}}
Authorization: Bearer {{accessToken}}
{
    "items": [
        {
            "order_item_id": {{orderItemId}},
            "product_id": {{productId}},
            "quantity": 3
        }
    ]
}
HTTP 200

GET {{baseUrl}}/api/order
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/{{orderId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: total revenue ──
GET {{baseUrl}}/api/order/monthly-total-revenue?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/yearly-total-revenue?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/merchant/monthly-total-revenue?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/merchant/yearly-total-revenue?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: revenue ──
GET {{baseUrl}}/api/order/monthly-revenue?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/yearly-revenue?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/merchant/monthly-revenue?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order/merchant/yearly-revenue?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/order/trashed/{{orderId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/order/restore/{{orderId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Trash lagi → restore/all → trash lagi → permanent/all
POST {{baseUrl}}/api/order/trashed/{{orderId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/order/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/order/trashed/{{orderId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/order/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200

# Order kedua — coverage DELETE /permanent/:id (sibling dari permanent/all).
# PENTING: tetap setelah permanent/all — service order error bila tidak ada trashed
# saat /all dipanggil, sedangkan per-id delete aman untuk record trashed.
POST {{baseUrl}}/api/order/create
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "cashier_id": {{cashierId}},
    "items": [
        {
            "product_id": {{productId}},
            "quantity": 1
        }
    ]
}
HTTP 200
[Captures]
orderId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/order/trashed/{{orderId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/order/permanent/{{orderId2}}
Authorization: Bearer {{accessToken}}
HTTP 200
