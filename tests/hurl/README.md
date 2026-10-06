# Hurl E2E — API Gateway (localhost:5000)

Suite e2e route POS untuk `service/apigateway` (Echo :5000). File lama
domain payment (`card.hurl`, `saldo.hurl`, `topup.hurl`, `transfer.hurl`,
`withdraw.hurl`) sudah **dihapus** (gap #9) — file di sini mengikuti
inventori route aktual (§5.4.1 SUPER_PLANNING.md).

## Prasyarat

- Stack berjalan: `just up` (atau service + gateway lokal).
- Seeder dijalankan agar data referensi (user, merchant, dsb.) ada:
  `just seeder`.
- Email test memakai `{{uuid}}` sehingga aman dijalankan berulang (data
  unik per run).

## Menjalankan

Semua file memakai variabel `baseUrl` (default tidak diset):

```bash
hurl --variable baseUrl=http://localhost:5000 --test tests/hurl/
```

Satu file saja:

```bash
hurl --variable baseUrl=http://localhost:5000 --test tests/hurl/auth.hurl
```

## Isi suite

| File | Cakupan |
|---|---|
| `auth.hurl` | hello → register → login → me → 401 → refresh-token → forgot-password → reset-password (invalid → 404) |
| `user.hurl` | lifecycle user + active/trashed + restore-all/permanent-all + 404 |
| `role.hurl` | lifecycle role + active/trashed/find-by-user + restore-all/permanent-all |
| `category.hurl` | lifecycle category + semua stats (total-pricing & pricing, merchant & mycategory) |
| `cashier.hurl` | create merchant → create cashier → lifecycle + semua stats sales |
| `merchant.hurl` | lifecycle merchant + update-status + active/trashed + restore-all/permanent-all |
| `merchant_document.hurl` | lifecycle document + update-status + regression DocumentID vs MerchantID |
| `product.hurl` | merchant + category + create product → lifecycle + filter merchant/category + restore-all/permanent-all |
| `order.hurl` | chain penuh → create order → lifecycle + semua stats revenue |
| `transaction.hurl` | chain penuh → create transaction → lifecycle + semua stats status/method |
| `order_item.hurl` | query order item (query only) |

Semua file **self-contained** (register user sendiri di awal), sehingga bisa
dijalankan independen.

## E2E otomatis (docker compose infra + service lokal)

`just e2e-hurl` menjalankan `tests/hurl/run_e2e.sh`:

1. Infra up via `docker compose -f deployments/local/docker-compose.infra.yml`
   (postgres, redis, kafka, dan observability tools — service Go TIDAK
   di-container, dijalankan lokal).
2. Tunggu postgres/redis/kafka → reset DB → `go run service/migrate` → seeder.
3. Build service lokal → start service lokal (terhubung kafka compose).
4. Jalankan setiap `*.hurl` satu per satu (dengan jeda antar file agar tidak
   kena rate limiter 20 rps gateway).

Manual:

```bash
just e2e-hurl                    # semuanya otomatis
# atau step-by-step:
just infra-up
just migrate && just seeder
just build && just services-up
hurl --variable baseUrl=http://localhost:5000 --test tests/hurl/
just services-down && just infra-down
```

## Catatan

- Order/transaction memakai chain: register → merchant → cashier →
  category → product → order → transaction. Setiap `[Captures]` mengambil
  id dari `$.data.id`; bila response mapper berubah, sesuaikan jsonpath.
- `transaction` menghitung ulang `amount` server-side (total + PPN 11%);
  request amount harus >= nilai tersebut (file sudah memakai nilai cukup).
- Status negatif: route mengembalikan `400` untuk id tidak valid dan `404`
  untuk id tidak ditemukan (lihat `user.hurl`).
