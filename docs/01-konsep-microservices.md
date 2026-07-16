# Konsep Dasar Microservices

## 1. Bahasa Sama atau Beda? (Polyglot Architecture)

Kemampuan pakai bahasa berbeda-beda per service itu **opsi, bukan kewajiban**. Banyak perusahaan sengaja menyamakan bahasa untuk semua service.

### Keuntungan bahasa SAMA (direkomendasikan untuk tim kecil / belajar)

- **Developer fleksibel** — anggota tim bisa pindah antar service tanpa belajar bahasa baru.
- **Bisa bikin shared library** — satu package untuk logging, validasi JWT, konfigurasi DB, dipakai semua service.
- **Belajar lebih cepat** — tidak pusing sintaks dan environment berbeda.

### Kapan pakai bahasa BEDA

Hanya jika ada kebutuhan spesifik yang bahasa tertentu selesaikan lebih baik:

| Service | Bahasa | Alasan |
|---------|--------|--------|
| Core API | Go / Rust | Performa, concurrency tinggi, hemat memori |
| AI/ML Service | Python | Ekosistem PyTorch, NumPy, Pandas terbaik |
| SSR / Backend-for-Frontend | Node.js (TypeScript) | Sinkron dengan tim frontend React/Next.js |

**Untuk proyek latihan: pakai satu bahasa yang paling dikuasai.** Fokus ke pemecahan database dan komunikasi antar-service.

## 2. Code Sharing Antar-Service (Bahasa Sama)

Trade-off: DRY vs tight coupling. Terlalu banyak berbagi kode merusak esensi microservices (tiap service harus bisa berdiri sendiri).

### Cara 1: Monorepo (paling praktis untuk belajar / tim kecil)

Semua service dalam satu repo Git, dipisah folder:

```
├── services/
│   ├── user-service/
│   └── order-service/
├── shared/
│   ├── middleware/   (validasi token JWT)
│   ├── models/       (struktur data/DTO bersama)
│   └── utils/        (helper enkripsi/hashing)
```

- **Go**: pakai Go Workspaces (`go.work`)
- **TypeScript/Node.js**: NPM/Yarn Workspaces atau Turborepo

### Cara 2: Repo terpisah + private package (skala produksi)

Repo khusus (mis. `core-shared-lib`), ditarik via Git dengan versioning (Git Tags):

```json
"dependencies": {
  "my-shared-package": "git+ssh://git@github.com:username/core-shared-lib.git#v1.0.0"
}
```

Go: set `GOPRIVATE`, lalu `go get github.com/username/core-shared-lib@v1.0.0`.

### Cara 3: Skema generator (khusus struktur data / DTO)

Tulis kontrak di file `.proto` (gRPC) atau OpenAPI/Swagger (REST) di tempat netral. Tiap service generate kodenya sendiri via compiler.

### ⚠️ Aturan emas

**Jangan pernah bagikan business logic di shared folder.**

- ✅ Boleh: verifikasi JWT, format log, koneksi DB.
- ❌ Tidak boleh: logika hitung diskon, validasi status user.

Alasan: business logic berubah → semua service harus deploy ulang → tujuan independent deployment hilang.

## 3. REST API vs gRPC

Analogi: REST = surat fisik yang mudah dibaca manusia (JSON); gRPC = panggilan telepon digital yang transmisi data biner.

| Fitur | REST API | gRPC |
|-------|----------|------|
| Format data | JSON / XML (teks) | Protocol Buffers (biner) |
| Protokol | HTTP/1.1 (umumnya) | HTTP/2 (wajib) |
| Kecepatan | Standar / cepat | Sangat cepat (hemat bandwidth) |
| Dukungan browser | Sangat baik (bawaan) | Terbatas (butuh gRPC-Web proxy) |
| Streaming | Satu arah (SSE) | Dua arah (client & server streaming) |
| Kontrak kode | Opsional (OpenAPI/Swagger) | Wajib (file `.proto`) |
| Paradigma | Resource-oriented (`POST /users`) | RPC (`CreateUser(request)`) |

### Kapan pakai yang mana

- **REST**: frontend → backend (browser support bawaan, praktis).
- **gRPC**: backend → backend antar microservice (kecepatan tinggi, latensi rendah, efisien untuk ribuan panggilan/detik).

Keduanya biasa dipakai bersamaan, bukan saling menggantikan.

## 4. Kontrak Data gRPC (`.proto`)

```protobuf
syntax = "proto3";

package user;

option go_package = "./pb";

service UserService {
  rpc GetUserByID (UserRequest) returns (UserResponse);
  rpc CreateUser (CreateUserRequest) returns (UserResponse);
}

message UserRequest {
  string id = 1;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
  int32 age = 3;
}

message UserResponse {
  string id = 1;
  string name = 2;
  string email = 3;
  int32 age = 4;
}
```

### Poin penting

1. **`syntax = "proto3";`** wajib di baris pertama.
2. **Angka `= 1, = 2`** adalah field number (tag), bukan value. gRPC mengirim angka ini, bukan nama field, agar hemat. **Sekali dipakai di production, jangan diubah** — merusak kompatibilitas data.
3. **Tipe data eksplisit**: `int32` (angka standar), `int64` (angka besar/ID unik).

### Generate kode

```bash
# install protoc + plugin bahasa (mis. protoc-gen-go), lalu:
protoc --go_out=. --go-grpc_out=. user.proto
```

Hasil: `user.pb.go` berisi struct dan fungsi gRPC otomatis. Tinggal isi logika bisnis.
