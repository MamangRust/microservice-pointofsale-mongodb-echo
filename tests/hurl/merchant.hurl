# Merchant E2E — lifecycle: create → update → update-status → trash → restore → delete permanent
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/merchant.hurl

POST {{baseUrl}}/api/auth/register
{
    "firstname": "Merchant",
    "lastname": "Owner",
    "email": "owner.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "owner.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Kopi {{uuid}}",
    "description": "Kopi nusantara",
    "address": "Jl. Merdeka No.1",
    "contact_email": "toko.{{uuid}}@example.com",
    "contact_phone": "081234567890",
    "status": "active"
}
HTTP 200
[Captures]
merchantId: jsonpath "$.data.id"

GET {{baseUrl}}/api/merchant
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/merchant/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant/update/{{merchantId}}
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Kopi Premium {{uuid}}",
    "description": "Kopi nusantara premium",
    "address": "Jl. Kopi No. 1",
    "contact_email": "kopi.{{uuid}}@example.com",
    "contact_phone": "081298765432",
    "status": "active"
}
HTTP 200

POST {{baseUrl}}/api/merchant/update-status/{{merchantId}}
Authorization: Bearer {{accessToken}}
{
    "status": "active"
}
HTTP 200

GET {{baseUrl}}/api/merchant/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/merchant/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant/trashed/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant/restore/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Trash lagi → restore/all → trash lagi → permanent/all
POST {{baseUrl}}/api/merchant/trashed/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant/trashed/{{merchantId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200

# Merchant kedua — coverage DELETE /permanent/:id (sibling dari permanent/all)
POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Perm {{uuid}}",
    "description": "Toko untuk test permanent",
    "address": "Jl. Perm No.9",
    "contact_email": "perm.{{uuid}}@example.com",
    "contact_phone": "081298765431",
    "status": "active"
}
HTTP 200
[Captures]
merchantId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/merchant/trashed/{{merchantId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/merchant/permanent/{{merchantId2}}
Authorization: Bearer {{accessToken}}
HTTP 200
