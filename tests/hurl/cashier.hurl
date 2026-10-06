# Cashier E2E — create merchant → create cashier → lifecycle
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/cashier.hurl

POST {{baseUrl}}/api/auth/register
{
    "firstname": "Cashier",
    "lastname": "Owner",
    "email": "cashier.owner.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "cashier.owner.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Kasir {{uuid}}",
    "description": "Toko untuk test cashier",
    "address": "Jl. Sudirman No.2",
    "contact_email": "kasir.{{uuid}}@example.com",
    "contact_phone": "081298765432",
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
    "name": "Kasir Utama {{uuid}}"
}
HTTP 201
[Captures]
cashierId: jsonpath "$.data.id"

GET {{baseUrl}}/api/cashier
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/{{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: total sales ──
GET {{baseUrl}}/api/cashier/monthly-total-sales?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/yearly-total-sales?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/merchant/monthly-total-sales?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/merchant/yearly-total-sales?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/mycashier/monthly-total-sales?year=2026&month=6&cashier_id={{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/mycashier/yearly-total-sales?year=2026&cashier_id={{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: sales ──
GET {{baseUrl}}/api/cashier/monthly-sales?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/yearly-sales?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/merchant/monthly-sales?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/merchant/yearly-sales?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/mycashier/monthly-sales?year=2026&month=6&cashier_id={{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/cashier/mycashier/yearly-sales?year=2026&cashier_id={{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/cashier/update/{{cashierId}}
Authorization: Bearer {{accessToken}}
{
    "name": "Kasir Senior {{uuid}}"
}
HTTP 200

POST {{baseUrl}}/api/cashier/trashed/{{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/cashier/restore/{{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/cashier/permanent/{{cashierId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Cashier kedua — coverage restore/all + permanent/all (sibling dari restore/:id + permanent/:id)
POST {{baseUrl}}/api/cashier/create
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "user_id": {{ownerId}},
    "name": "Kasir All {{uuid}}"
}
HTTP 201
[Captures]
cashierId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/cashier/trashed/{{cashierId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/cashier/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/cashier/trashed/{{cashierId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/cashier/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200
