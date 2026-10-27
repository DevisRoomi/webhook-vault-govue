# WEBHOOK VAULT // INGRESS LEDGER & RETRY DISPATCHER

```text
┌────────────────────────────────────────────────────────────────────────┐
│ WEBHOOK VAULT :: INGRESS LEDGER & HTTP DISPATCH RETRY CONTROLLER       │
│ Golang 1.27 RESTful API + JWT Auth + Vue 3 TS + TanStack Query v5      │
└────────────────────────────────────────────────────────────────────────┘
```

Platform teknis untuk menerima, mencatat log HTTP ingress, menginspeksi payload JSON mentah, dan mengeksekusi retry pengiriman webhook yang gagal (HTTP 500 / downstream timeout) secara real-time.

Dibangun dengan estetika **Archival Raw & Monospaced Wireframe**:
- Border kaku `1px` dengan radius `0px` pada seluruh tombol, input, dan panel.
- Font monospaced teknis (`JetBrains Mono`) dipadu font editorial serif (`Newsreader`) untuk deskripsi dokumentasi teknis.
- Layout asimetris 3 kolom: Channel Picker, Ingress Event Stream, dan Deep Inspector Pane.
- Indikator status raw brackets: `[200 OK]`, `[500 ERR]`, `[FORWARDED]`, `[LIVE]`.

---

## Arsitektur & Integrasi Sistem

- **Backend (Golang 1.27)**: RESTful API native routing (`net/http.ServeMux` Go 1.22+ pattern) pada port `http://127.0.0.1:8086`. Menangani ingestion, parsing header, logging database, dan eksekusi HTTP dispatcher/retry.
- **Database (FlyEnv MySQL 8.x)**: Port `3306`, database `webhook_vault`. Auto-migrasi skema (`users`, `endpoints`, `webhook_events`, `delivery_attempts`) dan data seed.
- **Frontend (Vue 3 + TypeScript + Vite)**: Desain modular atomic (`atoms`, `molecules`, `organisms`) menggunakan TanStack Query (`@tanstack/vue-query`) untuk smart caching dan auto-polling 5 detik.
- **Laravel Herd Gateway (`public/index.php`)**: Reverse proxy transparan yang meneruskan pemanggilan `/api/*` ke backend Golang port 8086 dan melayani single-page application di `http://webhook-vault-govue.test`.

---

## Fitur Utama

1. **Ingress Capture Gateway**:
   - Endpoint dinamis `/api/v1/ingest/{slug}` menerima metode `POST`, `PUT`, `GET`.
   - Menangkap seluruh header HTTP asli, source IP, query parameters, dan raw body JSON.
   - Perhitungan latensi dispatch dan verifikasi secret token (`X-Vault-Signature`, `X-Webhook-Secret`).

2. **Inspeksi Payload JSON & Audit Trail**:
   - Deep inspector menampilkan header HTTP, query URL, dan formatting JSON dengan 1-click clipboard copy.
   - Riwayat delivery attempt berurutan: attempt number, HTTP status response, downstream response headers, response body downstream, dan execution latency (ms).

3. **HTTP Dispatch Retry (Penanganan HTTP 500)**:
   - Fitur re-dispatch 1-klik untuk webhook yang gagal (`[⚡ EXECUTE HTTP RETRY DISPATCH]`).
   - Mendukung override URL tujuan (destination target) langsung dari UI inspector sebelum retry ditembakkan.
   - Pencatatan log retry baru tanpa menimpa riwayat kegagalan sebelumnya.

4. **Operasi Simulasi Webhook**:
   - Fitur pemicu webhook sintetis (`[SIMULATE 500 ERR]`) untuk menguji skenario deadlock downstream secara terkontrol.

---

## Panduan Menjalankan Sistem

### 1. Prasyarat
- FlyEnv / MySQL berjalan pada `127.0.0.1:3306` (kredensial bawaan: `root` / `root`)
- Laravel Herd aktif dengan parked folder `C:\Users\Lenovo\Herd`
- Go 1.22+ dan Bun / Node.js

### 2. Jalankan Backend Go
```bash
cd backend
go run main.go
# Atau jalankan binary yang sudah dikompilasi:
./vault-server.exe
```

### 3. Akses Antarmuka
- **Laravel Herd Site**: [http://webhook-vault-govue.test](http://webhook-vault-govue.test)
- **Direct API & Ingress**: `http://127.0.0.1:8086`

---

## Kredensial Akun Operator Demo

| Parameter | Nilai |
|---|---|
| **Username** | `dispatcher` |
| **Password** | `vault2026` |
| **Role** | `OPERATOR` |
| **Database** | `webhook_vault` (MySQL port 3306) |

*(Tersedia tombol `[AUTO-FILL: dispatcher]` pada modal sign-in untuk masuk instan).*

---

## Spesifikasi RESTful API

| Method | Endpoint | Akses | Deskripsi |
|---|---|---|---|
| `GET` | `/api/health` | Publik | Status operasional subsistem backend |
| `POST` | `/api/auth/login` | Publik | Autentikasi operator dan penerbitan JWT |
| `POST` | `/api/auth/register` | Publik | Inisialisasi operator baru |
| `GET` | `/api/auth/me` | Bearer JWT | Verifikasi sesi aktif operator |
| `GET` | `/api/stats` | Publik | Telemetri total ingress, 2xx, 5xx, dan channel aktif |
| `GET` | `/api/endpoints` | Publik | Daftar channel ingress terdaftar |
| `POST` | `/api/endpoints` | Bearer JWT | Registrasi channel ingress baru |
| `GET` | `/api/events` | Publik | Aliran log webhook dengan filter status & search |
| `GET` | `/api/events/{id}` | Publik | Detail lengkap event, payload, dan attempt trail |
| `POST` | `/api/events/{id}/retry` | Bearer JWT | Eksekusi retry pengiriman HTTP ke target |
| `POST` | `/api/test/simulate` | Bearer JWT | Memicu webhook sintetis (misal simulasi HTTP 500) |
| `POST/PUT` | `/api/v1/ingest/{slug}` | Publik / Token | Endpoint ingress penangkap webhook eksternal |

---

## Contoh Pengiriman Ingress via cURL

```bash
# Kirim payload webhook ke channel stripe-checkout
curl -X POST "http://webhook-vault-govue.test/api/v1/ingest/stripe-checkout" \
  -H "Content-Type: application/json" \
  -H "User-Agent: Stripe/1.0" \
  -d '{
    "id": "evt_live_test_001",
    "type": "payment_intent.succeeded",
    "amount": 2500000,
    "currency": "idr"
  }'
```

---

## Lisensi

MIT License © DevisRoomi
