# Order Item E2E — query only (RPC command tidak ada di order_item service)
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/order_item.hurl

# Register + login (self-contained)
POST {{baseUrl}}/api/auth/register
{
    "firstname": "OrderItem",
    "lastname": "Viewer",
    "email": "orderitem.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200

POST {{baseUrl}}/api/auth/login
{
    "email": "orderitem.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

# Find all order items
GET {{baseUrl}}/api/order-item
Authorization: Bearer {{accessToken}}
HTTP 200

# Find active / trashed
GET {{baseUrl}}/api/order-item/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/order-item/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# Find by order id (1 = hasil seeder, kosong bila tidak ada — HTTP 200)
GET {{baseUrl}}/api/order-item/1
Authorization: Bearer {{accessToken}}
HTTP 200
