# Coffee Shop POS API

RESTful API untuk sistem Point of Sale (POS) kedai kopi, dibangun dengan Go.

## Deskripsi

Coffee Shop POS API adalah backend service yang mengelola operasional kedai kopi, termasuk manajemen menu, pesanan, pembayaran via Midtrans, autentikasi JWT, dan caching Redis.

## Tech Stack

- **Language**: Go
- **Database**: MySQL
- **Cache**: Redis
- **Payment**: Midtrans
- **Auth**: JWT

## Getting Started

1. Copy environment variables:
   ```bash
   cp .env.example .env
   ```

2. Sesuaikan nilai variabel di `.env`

3. Jalankan aplikasi:
   ```bash
   go run cmd/api/main.go
   ```

## Struktur Folder

```
coffee-pos/
├── cmd/api/          # Entry point aplikasi
├── config/           # Konfigurasi aplikasi
├── internal/
│   ├── entity/       # Domain struct
│   ├── repository/   # Repository interfaces & implementasi
│   ├── service/      # Business logic
│   ├── handler/      # HTTP handlers
│   ├── middleware/   # HTTP middleware
│   └── dto/          # Request/response struct
├── migrations/       # SQL migration files
└── pkg/
    ├── database/     # Koneksi database
    ├── redis/        # Koneksi Redis
    ├── jwt/          # JWT helper
    ├── response/     # Response helper
    └── validator/    # Input validator
```
