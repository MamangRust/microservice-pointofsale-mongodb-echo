# Product E2E — create merchant → create category → create product → lifecycle
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/product.hurl

POST {{baseUrl}}/api/auth/register
{
    "firstname": "Product",
    "lastname": "Owner",
    "email": "product.owner.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "product.owner.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Produk {{uuid}}",
    "description": "Toko untuk test produk",
    "address": "Jl. Produk No.3",
    "contact_email": "produk.{{uuid}}@example.com",
    "contact_phone": "081298765431",
    "status": "active"
}
HTTP 200
[Captures]
merchantId: jsonpath "$.data.id"

POST {{baseUrl}}/api/category/create
Authorization: Bearer {{accessToken}}
{
    "name": "Kopi-{{uuid}}",
    "description": "Kategori kopi"
}
HTTP 201
[Captures]
categoryId: jsonpath "$.data.id"
categoryName: jsonpath "$.data.name"

POST {{baseUrl}}/api/product/create
Authorization: Bearer {{accessToken}}
[MultipartFormData]
merchant_id: {{merchantId}}
category_id: {{categoryId}}
name: Kopi Susu {{uuid}}
description: Es kopi susu gula aren
price: 25000
count_in_stock: 100
brand: Nusantara
weight: 250
image_product: file, fixtures/product.jpg; image/jpeg
HTTP 201
[Captures]
productId: jsonpath "$.data.id"

GET {{baseUrl}}/api/product
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/product/{{productId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Filter by merchant / category + active / trashed
GET {{baseUrl}}/api/product/merchant/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Nama kategori memakai slug (tanpa spasi) — filter urlEncode hurl 8.x
# menghasilkan URL malformed bila nilai mengandung spasi (libcurl error 3).
GET {{baseUrl}}/api/product/category/{{categoryName}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/product/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/product/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/product/update/{{productId}}
Authorization: Bearer {{accessToken}}
[MultipartFormData]
merchant_id: {{merchantId}}
category_id: {{categoryId}}
name: Kopi Susu Premium {{uuid}}
description: Es kopi susu gula aren premium
price: 30000
count_in_stock: 90
brand: Nusantara
weight: 250
HTTP 200

POST {{baseUrl}}/api/product/trashed/{{productId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/product/restore/{{productId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Trash lagi → restore/all → trash lagi → permanent/all
POST {{baseUrl}}/api/product/trashed/{{productId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/product/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/product/trashed/{{productId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/product/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200

# Product kedua — coverage DELETE /permanent/:id (sibling dari permanent/all)
POST {{baseUrl}}/api/product/create
Authorization: Bearer {{accessToken}}
[MultipartFormData]
merchant_id: {{merchantId}}
category_id: {{categoryId}}
name: Kopi Perm {{uuid}}
description: Produk untuk test permanent
price: 15000
count_in_stock: 20
brand: Nusantara
weight: 100
image_product: file, fixtures/product.jpg; image/jpeg
HTTP 201
[Captures]
productId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/product/trashed/{{productId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/product/permanent/{{productId2}}
Authorization: Bearer {{accessToken}}
HTTP 200
