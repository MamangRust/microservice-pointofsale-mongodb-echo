# Auth E2E — register → login → me
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/auth.hurl

# Hello (public)
GET {{baseUrl}}/api/auth/hello
HTTP 200

# Register (email unik per run)
POST {{baseUrl}}/api/auth/register
{
    "firstname": "John",
    "lastname": "Doe",
    "email": "john.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
userId: jsonpath "$.data.id"

# Login
POST {{baseUrl}}/api/auth/login
{
    "email": "john.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"
refreshToken: jsonpath "$.data.refresh_token"

# Me (authorized)
GET {{baseUrl}}/api/auth/me
Authorization: Bearer {{accessToken}}
HTTP 200

# Unauthorized without token
GET {{baseUrl}}/api/auth/me
HTTP 401

# Refresh token
POST {{baseUrl}}/api/auth/refresh-token
{
    "refresh_token": "{{refreshToken}}"
}
HTTP 200

# Forgot password (email terdaftar)
POST {{baseUrl}}/api/auth/forgot-password
{
    "email": "john.{{uuid}}@example.com"
}
HTTP 200

# Reset password dengan token tidak valid → 404 (tidak membocorkan info)
POST {{baseUrl}}/api/auth/reset-password
{
    "reset_token": "invalid-token",
    "password": "newpassword123",
    "confirm_password": "newpassword123"
}
HTTP 404

# Verify code — route publik (tanpa auth); kode tidak dikenal → 404
GET {{baseUrl}}/api/auth/verify-code?verify_code=invalid-code-{{uuid}}
HTTP 404
