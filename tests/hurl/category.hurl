# Category E2E — lifecycle + semua stats endpoints
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/category.hurl

# Register admin (self-contained)
POST {{baseUrl}}/api/auth/register
{
    "firstname": "Category",
    "lastname": "Admin",
    "email": "category.admin.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "category.admin.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

# Merchant (untuk stats by merchant)
POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Kategori {{uuid}}",
    "description": "Toko untuk stats kategori",
    "address": "Jl. Kategori No.7",
    "contact_email": "kategori.{{uuid}}@example.com",
    "contact_phone": "081298765427",
    "status": "active"
}
HTTP 200
[Captures]
merchantId: jsonpath "$.data.id"

POST {{baseUrl}}/api/category/create
Authorization: Bearer {{accessToken}}
{
    "name": "Minuman {{uuid}}",
    "description": "Kategori minuman segar"
}
HTTP 201
[Captures]
categoryId: jsonpath "$.data.id"

GET {{baseUrl}}/api/category
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/{{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: total pricing ──
GET {{baseUrl}}/api/category/monthly-total-pricing?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/yearly-total-pricing?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/merchant/monthly-total-pricing?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/merchant/yearly-total-pricing?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/mycategory/monthly-total-pricing?year=2026&month=6&category_id={{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/mycategory/yearly-total-pricing?year=2026&category_id={{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Stats: pricing ──
GET {{baseUrl}}/api/category/monthly-pricing?year=2026&month=6
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/yearly-pricing?year=2026
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/merchant/monthly-pricing?year=2026&month=6&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/merchant/yearly-pricing?year=2026&merchant_id={{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/mycategory/monthly-pricing?year=2026&month=6&category_id={{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/category/mycategory/yearly-pricing?year=2026&category_id={{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# ── Lifecycle ──
POST {{baseUrl}}/api/category/update/{{categoryId}}
Authorization: Bearer {{accessToken}}
{
    "name": "Minuman Segar {{uuid}}",
    "description": "Kategori minuman segar updated"
}
HTTP 200

POST {{baseUrl}}/api/category/trashed/{{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/category/restore/{{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/category/trashed/{{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/category/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/category/trashed/{{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/category/permanent/{{categoryId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Category kedua — coverage POST /permanent/all (sibling dari DELETE /permanent/:id)
POST {{baseUrl}}/api/category/create
Authorization: Bearer {{accessToken}}
{
    "name": "Minuman All {{uuid}}",
    "description": "Kategori untuk test permanent/all"
}
HTTP 201
[Captures]
categoryId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/category/trashed/{{categoryId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/category/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200
