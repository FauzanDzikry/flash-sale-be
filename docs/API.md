# Dokumentasi API E-commerce Flash Sale

## 1. Pendahuluan

### Deskripsi

API ini merupakan backend untuk layanan **E-commerce Flash Sale**. Saat ini menyediakan health check dan manajemen autentikasi: registrasi pengguna, login, logout, serta pengambilan profil pengguna. Endpoint terkait produk dan flash sale dapat ditambahkan pada versi mendatang.

### Tujuan

Dokumentasi ini memungkinkan integrasi oleh client (web/mobile) dan pengembang eksternal dengan kontrak yang jelas: base URL, autentikasi, format respons, kode kesalahan, serta detail setiap endpoint.

---

## 2. Base URL

Semua endpoint API berada di bawah base path `/api/v1`.

| Lingkungan   | Base URL                              |
|-------------|----------------------------------------|
| Development | `http://localhost:8080/api/v1`        |
| Production  | `https://api.example.com/api/v1`       |

Ganti host/port sesuai environment Anda. Server default Gin menggunakan port `8080` jika tidak dikonfigurasi lain.

---

## 3. Autentikasi

### Metode

API menggunakan **Bearer Token (JWT)**. Request yang memerlukan autentikasi harus menyertakan header:

```
Authorization: Bearer <access_token>
```

### Mendapatkan Token

Panggil endpoint **POST** `/api/v1/auth/login` dengan body JSON berisi `email` dan `password`. Response sukses berisi `access_token`, `token_type`, `expires_in` (detik), dan data `user`. Gunakan nilai `access_token` sebagai `<access_token>` di header di atas.

### Endpoint yang Dilindungi

Hanya endpoint berikut yang memerlukan header `Authorization: Bearer <access_token>`:

- **GET** `/api/v1/auth/me` — mengambil profil user saat ini

Endpoint `register`, `login`, dan `logout` tidak memerlukan token.

### Contoh Request dengan Token

```bash
curl -X GET "http://localhost:8080/api/v1/auth/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Alur Autentikasi

```mermaid
sequenceDiagram
  participant Client
  participant API
  participant Auth

  Client->>API: POST /auth/register (email, password, name)
  API->>Auth: Register
  Auth-->>API: UserResponse
  API-->>Client: 201 User

  Client->>API: POST /auth/login (email, password)
  API->>Auth: Login
  Auth-->>API: LoginResponse (access_token, user)
  API-->>Client: 200 + access_token

  Client->>API: GET /auth/me + Authorization: Bearer token
  API->>API: JWT middleware
  API->>Auth: GetProfile(user_id)
  Auth-->>API: UserResponse
  API-->>Client: 200 User
```

---

## 4. Standar Respons

### Format Data

- Semua respons dalam format **JSON**.
- Tidak ada envelope global: respons sukses mengembalikan object/array langsung (misalnya object user atau object login).
- Content-Type: `application/json`.

### Skema Tanggal

Untuk field bertipe datetime (bila ada di endpoint mendatang), gunakan **ISO 8601**, contoh: `2025-02-24T10:00:00Z`.

### Pagination

Pagination saat ini belum diimplementasikan. Dokumentasi pagination (parameter `page`, `limit`, serta format respons berisi `data`, `total`, `page`) akan ditambahkan ketika endpoint yang mendukung list tersedia.

---

## 5. Kode Kesalahan (Error Handling)

Respons error berupa JSON dengan field `message` (wajib). Untuk validasi (400), field `error` berisi detail dari validator.

| Kode HTTP | Situasi | Bentuk Respons |
|-----------|---------|-----------------|
| 400 | Validasi request gagal (binding) | `{"message": "Invalid request", "error": "<detail>"}` |
| 401 | Header Authorization kosong | `{"message": "Unauthorized header is required"}` |
| 401 | Format Authorization salah (bukan Bearer) | `{"message": "Invalid authorization format"}` |
| 401 | Token tidak valid atau kedaluwarsa | `{"message": "Invalid token"}` |
| 401 | Token sudah di-revoke (setelah logout) | `{"message": "Token has been revoked"}` |
| 401 | Email atau password salah (login) | `{"message": "Invalid email or password"}` |
| 403 | Reserved; belum digunakan | - |
| 404 | User tidak ditemukan (mis. GET /auth/me) | `{"message": "User not found"}` |
| 409 | Email sudah terdaftar (register) | `{"message": "Email already registered"}` |
| 500 | Kesalahan server (register/login gagal, invalid context) | `{"message": "..."}` |

---

## 6. Dokumentasi Endpoint

### 6.1 Health Check

**GET** `/api/v1/ping`

Memeriksa koneksi ke aplikasi. Tidak memerlukan autentikasi.

#### Parameter

Tidak ada parameter path, query, atau body.

#### Contoh Request

```bash
curl -X GET "http://localhost:8080/api/v1/ping"
```

#### Response Sukses (200)

```json
{
  "status": "ok",
  "message": "pong"
}
```

#### Response Error

Endpoint ini tidak mengembalikan error standar; digunakan untuk health check saja.

---

### 6.2 Registrasi User

**POST** `/api/v1/auth/register`

Mendaftarkan user baru. Tidak memerlukan autentikasi.

#### Parameter (Body, JSON)

| Parameter | Tipe   | Required | Deskripsi                          |
|-----------|--------|----------|------------------------------------|
| email     | string | Required | Alamat email; format email valid   |
| password  | string | Required | Minimal 8 karakter                 |
| name      | string | Required | Nama user                          |

#### Contoh Request

```bash
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

#### Response Sukses (201)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe"
}
```

#### Response Error (400)

```json
{
  "message": "Invalid request",
  "error": "Key: 'RegisterRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"
}
```

#### Response Error (409)

```json
{
  "message": "Email already registered"
}
```

#### Response Error (500)

```json
{
  "message": "Failed to register"
}
```

---

### 6.3 Login

**POST** `/api/v1/auth/login`

Login dengan email dan password. Mengembalikan JWT (`access_token`) dan data user. Tidak memerlukan autentikasi.

#### Parameter (Body, JSON)

| Parameter | Tipe   | Required | Deskripsi                |
|-----------|--------|----------|--------------------------|
| email     | string | Required | Alamat email             |
| password  | string | Required | Minimal 8 karakter       |

#### Contoh Request

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Response Sukses (200)

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

#### Response Error (400)

```json
{
  "message": "Invalid request",
  "error": "Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

#### Response Error (401)

```json
{
  "message": "Invalid email or password"
}
```

#### Response Error (500)

```json
{
  "message": "Failed to login"
}
```

---

### 6.4 Logout

**POST** `/api/v1/auth/logout`

Logout. Secara server-side tidak melakukan invalidasi token; klien disarankan membuang token. Tidak memerlukan autentikasi.

#### Parameter

Tidak ada parameter path, query, atau body.

#### Contoh Request

```bash
curl -X POST "http://localhost:8080/api/v1/auth/logout" \
  -H "Content-Type: application/json"
```

#### Response Sukses (200)

```json
{
  "message": "Logged out successfully"
}
```

---

### 6.5 Profil User (Me)

**GET** `/api/v1/auth/me`

Mengambil profil user yang sedang login. **Memerlukan** header `Authorization: Bearer <access_token>`.

#### Parameter

Tidak ada parameter path, query, atau body. User diidentifikasi dari JWT.

#### Contoh Request

```bash
curl -X GET "http://localhost:8080/api/v1/auth/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response Sukses (200)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe"
}
```

#### Response Error (401)

```json
{
  "message": "Unauthorized header is required"
}
```

atau

```json
{
  "message": "Invalid authorization format"
}
```

atau

```json
{
  "message": "Invalid token"
}
```

atau (token sudah logout/revoke)

```json
{
  "message": "Token has been revoked"
}
```

#### Response Error (404)

```json
{
  "message": "User not found"
}
```

#### Response Error (500)

```json
{
  "message": "Invalid User Context"
}
```

---

## 7. Rate Limiting

Rate limiting saat ini **tidak diimplementasikan**. Batas request per menit/jam serta header respons (misalnya `X-RateLimit-Limit`, `X-RateLimit-Remaining`) akan didokumentasikan jika fitur tersebut ditambahkan di kemudian hari.
