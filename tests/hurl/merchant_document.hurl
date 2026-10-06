# Merchant Document E2E — lifecycle + regression DocumentID vs MerchantID
# Jalankan: hurl --variable baseUrl=http://localhost:5000 tests/hurl/merchant_document.hurl

# Register owner
POST {{baseUrl}}/api/auth/register
{
    "firstname": "Doc",
    "lastname": "Owner",
    "email": "doc.owner.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200
[Captures]
ownerId: jsonpath "$.data.id"

POST {{baseUrl}}/api/auth/login
{
    "email": "doc.owner.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

POST {{baseUrl}}/api/merchant/create
Authorization: Bearer {{accessToken}}
{
    "user_id": {{ownerId}},
    "name": "Toko Dokumen {{uuid}}",
    "description": "Toko untuk test dokumen",
    "address": "Jl. Dokumen No.6",
    "contact_email": "dokumen.{{uuid}}@example.com",
    "contact_phone": "081298765428",
    "status": "active"
}
HTTP 200
[Captures]
merchantId: jsonpath "$.data.id"

# Document 1
POST {{baseUrl}}/api/merchant-documents/create
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "document_type": "license",
    "document_url": "https://example.com/license.pdf"
}
HTTP 200
[Captures]
documentId: jsonpath "$.data.id"

# Document 2 — ID-nya beda dari merchant_id; update harus menarget DocumentID
POST {{baseUrl}}/api/merchant-documents/create
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "document_type": "id_card",
    "document_url": "https://example.com/id.pdf"
}
HTTP 200
[Captures]
documentId2: jsonpath "$.data.id"

GET {{baseUrl}}/api/merchant-documents
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/merchant-documents/{{documentId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/merchant-documents/active
Authorization: Bearer {{accessToken}}
HTTP 200

# Update dokumen 2 — field status/note ikut dikirim (diteruskan API → gRPC)
POST {{baseUrl}}/api/merchant-documents/updates/{{documentId2}}
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "document_type": "id_card_updated",
    "document_url": "https://example.com/id-updated.pdf",
    "status": "pending",
    "note": "reupload"
}
HTTP 200

# Update status dokumen 2
POST {{baseUrl}}/api/merchant-documents/update-status/{{documentId2}}
Authorization: Bearer {{accessToken}}
{
    "merchant_id": {{merchantId}},
    "status": "verified",
    "note": "Approved"
}
HTTP 200

# Trash → list trashed → restore (dokumen 1)
POST {{baseUrl}}/api/merchant-documents/trashed/{{documentId}}
Authorization: Bearer {{accessToken}}
HTTP 200

GET {{baseUrl}}/api/merchant-documents/trashed
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant-documents/restore/{{documentId}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Delete permanent dokumen 2
DELETE {{baseUrl}}/api/merchant-documents/permanent/{{documentId2}}
Authorization: Bearer {{accessToken}}
HTTP 200

# Coverage restore/all + permanent/all — dokumen 1 di-trash dulu (sibling dari restore/:id + permanent/:id)
POST {{baseUrl}}/api/merchant-documents/trashed/{{documentId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant-documents/restore/all
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant-documents/trashed/{{documentId}}
Authorization: Bearer {{accessToken}}
HTTP 200

POST {{baseUrl}}/api/merchant-documents/permanent/all
Authorization: Bearer {{accessToken}}
HTTP 200
