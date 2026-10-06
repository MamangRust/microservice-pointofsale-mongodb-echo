# User E2E — lifecycle + all query endpoints
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/user.hurl

# Register admin (self-contained)
POST {{baseUrl}}/api/auth/register
{
    "firstname": "User",
    "lastname": "Admin",
    "email": "user.admin.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
adminId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "user.admin.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

# Create user
POST {{baseUrl}}/api/user/create
Authorization: Bearer {{accessToken}}
{
    "firstname": "Jane",
    "lastname": "Doe",
    "email": "jane.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
newUserId: jsonpath "$.data.id"

# Find all / active / trashed
GET {{baseUrl}}/api/user
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/user/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/user/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# Find by id
GET {{baseUrl}}/api/user/{{newUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Update user
POST {{baseUrl}}/api/user/update/{{newUserId}}
Authorization: Bearer {{accessToken}}
{
    "firstname": "Jane",
    "lastname": "Smith",
    "email": "jane.smith.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200

# Invalid id → 404 (tidak ditemukan)
GET {{baseUrl}}/api/user/-1
Authorization: Bearer {{accessToken}}
HTTP 404

# Not found id → 404
GET {{baseUrl}}/api/user/99999999
Authorization: Bearer {{accessToken}}
HTTP 404

# Trash → restore → delete permanent
POST {{baseUrl}}/api/user/trashed/{{newUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/user/restore/{{newUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Trash lagi → restore/all → trash lagi → permanent/all
POST {{baseUrl}}/api/user/trashed/{{newUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/user/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/user/trashed/{{newUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/user/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200

# User kedua — coverage DELETE /permanent/:id (sibling dari permanent/all)
POST {{baseUrl}}/api/user/create
Authorization: Bearer {{accessToken}}
{
    "firstname": "Perm",
    "lastname": "User",
    "email": "perm.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
permUserId: jsonpath "$.data.id"

POST {{baseUrl}}/api/user/trashed/{{permUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/user/permanent/{{permUserId}}
Authorization: Bearer {{accessToken}}
HTTP 200
