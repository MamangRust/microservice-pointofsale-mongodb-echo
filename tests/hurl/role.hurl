# Role E2E — lifecycle + all query endpoints
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/role.hurl

# Register admin (self-contained)
POST {{baseUrl}}/api/auth/register
{
    "firstname": "Role",
    "lastname": "Admin",
    "email": "role.admin.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
adminId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "role.admin.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/role
Authorization: Bearer {{accessToken}}
{
    "name": "cashier-{{uuid}}"
}
HTTP 200
[Captures]
roleId: jsonpath "$.data.id"

GET {{baseUrl}}/api/role
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/role/{{roleId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/role/active
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/role/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

# Roles milik user (ROLE_ADMIN dari register)
GET {{baseUrl}}/api/role/user/{{adminId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/role/update/{{roleId}}
Authorization: Bearer {{accessToken}}
{
    "name": "upd-{{uuid}}"
}
HTTP 200

POST {{baseUrl}}/api/role/trashed/{{roleId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/role/restore/{{roleId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Trash lagi → restore/all → trash lagi → permanent-all
POST {{baseUrl}}/api/role/trashed/{{roleId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/role/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/role/trashed/{{roleId}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/role/permanent-all
Authorization: Bearer {{accessToken}}
HTTP 200

# Role kedua — coverage DELETE /permanent/:id (route swagger; permanent-all sudah dicover)
POST {{baseUrl}}/api/role
Authorization: Bearer {{accessToken}}
{
    "name": "perm-{{uuid}}"
}
HTTP 200
[Captures]
roleId2: jsonpath "$.data.id"

POST {{baseUrl}}/api/role/trashed/{{roleId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

DELETE {{baseUrl}}/api/role/permanent/{{roleId2}}
Authorization: Bearer {{accessToken}}
HTTP 200
