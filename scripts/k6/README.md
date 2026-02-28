# Grafana k6 - Load & Race Condition Tests

## Prerequisites

1. **Install Grafana k6** (berbeda dari Grafana dashboard):
   - Windows: `winget install k6` atau [download](https://github.com/grafana/k6/releases)
   - Atau: `choco install k6`
2. Server berjalan (Postgres, Redis, API) pada `http://localhost:8080`

## Checkout Race Condition Test

 Mensimulasikan banyak pengguna bersamaan memanggil checkout untuk produk yang sama dengan stok terbatas.

### Menjalankan

```bash
# Default (BASE_URL=http://localhost:8080)
k6 run scripts/k6/checkout_race.js

# Custom base URL
k6 run -e BASE_URL=http://localhost:8080 scripts/k6/checkout_race.js

# Override VUs & duration (via CLI)
k6 run --vus 100 --duration 60s scripts/k6/checkout_race.js
```

### Scenario

- **Setup**: Register user → Login → Create product (stock 15)
- **Load**: Ramp 0→30 VUs (5s), 30→50 VUs (15s), 50→80 VUs (10s)
- Setiap VU mengirim `POST /api/v1/checkouts` dengan `quantity: 2`
- **Expected**: 202 (accepted) atau 400 (insufficient stock); tidak ada 5xx

### Custom Metrics

| Metric              | Description                    |
|---------------------|--------------------------------|
| `checkout_accepted` | Request diterima (202)         |
| `checkout_rejected` | Stok tidak cukup (400)        |
| `checkout_errors`   | Error lain (4xx/5xx)          |
| `checkout_duration` | Latency per request            |

### Verifikasi Race Condition

Setelah test, cek di database:

- `stock` tidak pernah negatif
- `SUM(checkouts.quantity) WHERE product_id = ?` ≤ stok awal produk

```sql
-- Contoh verifikasi
SELECT p.id, p.stock, COALESCE(SUM(c.quantity), 0) AS total_sold
FROM products p
LEFT JOIN checkouts c ON c.product_id = p.id
WHERE p.name LIKE 'Flash Product Race%'
GROUP BY p.id, p.stock;
```
