# F5 — Stats pipeline E2E (ClickHouse via stats-reader).
#
# Alur data: service domain --(outbox)--> Kafka stats.pos.*.event -->
# stats-writer --> ClickHouse --> stats-reader (gRPC) --> apigateway.
#
# Dependensi run_e2e.sh sebelum file ini:
#   1. stats_writer + stats_reader ikut di-start
#   2. `stats-writer backfill` dijalankan (seed OLTP -> ClickHouse)
#   3. jeda >= 1 flush interval (5s) setelah suite domain, supaya event live
#      order/transaction dari order.hurl & transaction.hurl sudah termaterialisasi
#
# Jalankan manual:
#   hurl --test --variable baseUrl=http://localhost:5000 \
#     --variable uuid=abc123 --variable year=2026 --variable month=8 \
#     --variable cashier_id=1 tests/hurl/stats.hurl

# ── Auth: register + login (token untuk semua request) ─────────────────────
POST {{baseUrl}}/api/auth/register
{
    "firstname": "Stats",
    "lastname": "Tester",
    "email": "stats.{{uuid}}@example.com",
    "password": "password123",
    "confirm_password": "password123"
}
HTTP 200

POST {{baseUrl}}/api/auth/login
{
    "email": "stats.{{uuid}}@example.com",
    "password": "password123"
}
HTTP 200
[Captures]
accessToken: jsonpath "$.data.access_token"

# ── Order stats ─────────────────────────────────────────────────────────────
GET {{baseUrl}}/api/order-stats/monthly-total-revenue?year={{year}}&month={{month}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET {{baseUrl}}/api/order-stats/yearly-revenue?year={{year}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET {{baseUrl}}/api/order-stats/cashier/monthly-revenue?cashier_id={{cashier_id}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

# ── Product stats ───────────────────────────────────────────────────────────
GET {{baseUrl}}/api/product-stats/monthly-sold?year={{year}}&month={{month}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

# ── Category stats ──────────────────────────────────────────────────────────
GET {{baseUrl}}/api/category-stats/monthly-sold?year={{year}}&month={{month}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

# ── Transaction stats ───────────────────────────────────────────────────────
GET {{baseUrl}}/api/transaction-stats/monthly-success?year={{year}}&month={{month}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

# ── Cashier stats ───────────────────────────────────────────────────────────
GET {{baseUrl}}/api/cashier-stats/monthly-orders?cashier_id={{cashier_id}}
Authorization: Bearer {{accessToken}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

# ── Error contract: param invalid -> 400 ────────────────────────────────────
GET {{baseUrl}}/api/order-stats/monthly-total-revenue?year=0&month=13
Authorization: Bearer {{accessToken}}
HTTP 400
[Asserts]
jsonpath "$.status" == "error"
