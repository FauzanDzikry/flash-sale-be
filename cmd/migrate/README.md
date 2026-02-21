# Migration Runner

Runner untuk menjalankan migration database (PostgreSQL) dari file SQL di folder `migrations/`. Satu file `main.go` dipanggil dengan flag untuk **up**, **down**, **force**, atau cek versi.

## Persyaratan

- Jalankan dari **root project** (folder yang berisi `go.mod` dan folder `migrations/`).
- Database URL harus tersedia via env **`DATABASE_URL`** atau flag **`-dsn`**.
- File **`.env`** di root project akan otomatis dimuat (package `godotenv`). Nilai di `.env` mengisi env sebelum flag dibaca.

Format URL PostgreSQL:

```
postgres://user:password@host:port/database?sslmode=disable
```

## Flag

| Flag | Deskripsi |
|------|------------|
| `-up` | Jalankan semua migration yang belum dijalankan (migrate up). |
| `-down` | Jalankan migration ke bawah. Default 1 step. Gunakan `-steps N` untuk N step. |
| `-down-all` | Rollback semua migration sampai versi 0. |
| `-steps N` | Jumlah migration yang dijalankan. Bersama `-up`: N migration ke atas; bersama `-down`: N migration ke bawah. `0` = semua (hanya untuk `-up`). |
| `-force V` | Paksa set versi schema ke `V`. Berguna saat migration dalam keadaan *dirty* (gagal di tengah). |
| `-dsn "URL"` | URL koneksi database (override env `DATABASE_URL`). |
| `-migrations-dir "path"` | Path ke folder migration (override env `MIGRATIONS_DIR`). Default: `migrations`. |

Tanpa flag: program hanya menampilkan **versi migration saat ini** (dan status dirty).

## Cara Menggunakan

### 1. Set variabel lingkungan (opsional)

Anda bisa menulis `DATABASE_URL` dan `MIGRATIONS_DIR` di file **`.env`** di root project; runner akan memuatnya otomatis. Atau set manual:

**Windows (CMD):**

```cmd
set DATABASE_URL=postgres://postgres:password@localhost:5432/flash_sale?sslmode=disable
```

**Windows (PowerShell):**

```powershell
$env:DATABASE_URL="postgres://postgres:password@localhost:5432/flash_sale?sslmode=disable"
```

**Linux / macOS:**

```bash
export DATABASE_URL="postgres://postgres:password@localhost:5432/flash_sale?sslmode=disable"
```

### 2. Jalankan dengan `go run`

Dari root project:

```bash
# Cek versi migration saat ini
go run ./cmd/migrate

# Jalankan semua migration ke atas
go run ./cmd/migrate -up

# Jalankan 1 migration ke atas
go run ./cmd/migrate -up -steps 1

# Jalankan 1 migration ke bawah
go run ./cmd/migrate -down -steps 1

# Rollback semua migration (pakai satu flag: -down-all, bukan -down -all)
go run ./cmd/migrate -down-all

# Paksa set versi (misal setelah dirty)
go run ./cmd/migrate -force 0
```

### 3. Pakai DSN lewat flag (tanpa env)

```bash
go run ./cmd/migrate -dsn "postgres://user:pass@localhost:5432/mydb?sslmode=disable" -up
```

### 4. Pakai folder migration lain

```bash
go run ./cmd/migrate -migrations-dir ./custom_migrations -up
```

Atau dengan env:

```bash
set MIGRATIONS_DIR=./custom_migrations
go run ./cmd/migrate -up
```

### 5. Build binary lalu jalankan

```bash
go build -o migrate.exe ./cmd/migrate
.\migrate.exe -up
.\migrate.exe -down -steps 1
.\migrate.exe
```

## Folder migration

File migration berada di folder **`migrations/`** (di root project) dengan format:

- `000001_nama_migration.up.sql`   — dijalankan saat **up**
- `000001_nama_migration.down.sql` — dijalankan saat **down**

Nomor versi harus berurutan (000001, 000002, …). Tambah file baru dengan nomor berikutnya saat menambah migration.

## Menangani dirty state

Jika migration gagal di tengah jalan, schema bisa dalam keadaan *dirty*. Untuk memperbaiki:

1. Perbaiki database atau file SQL sesuai kebutuhan.
2. Paksa versi ke nilai yang sesuai, misalnya ke versi sebelumnya atau 0:
   ```bash
   go run ./cmd/migrate -force 0
   ```
3. Jalankan lagi `-up` atau `-down` sesuai kebutuhan.

## Contoh output

**Versi saat ini (tanpa flag):**

```
current version: 2 (dirty: false)
```

**Migrate up berhasil:**

```
migrate up ok
```

**Sudah up to date:**

```
no change (already up to date)
```

**Down all berhasil:**

```
migrate down all ok (2 step(s))
```
