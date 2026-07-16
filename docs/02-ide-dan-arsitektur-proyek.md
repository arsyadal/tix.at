# Ide Proyek & Perbandingan Kerumitan

## Ide Proyek Utama

### 1. Sistem Booking Tiket Konser (fokus: concurrency & caching) ⭐ dipilih untuk tix.at

Simulasi "war tiket" — memaksa mikirkan race condition dan kecepatan respons.

**Service:**
- **Event/Ticket Service** — data konser + kuota tiket.
- **Booking Service** — antrean request pemesanan.
- **Payment Simulator** — simulasi pembayaran (tanpa payment gateway asli).
- **Notification Service** — simulasi email konfirmasi.

**Alur:**
1. Frontend → API Gateway → daftar konser.
2. Klik "Beli" → Booking Service cek sisa tiket di **Redis** (bukan DB utama, supaya cepat).
3. Tiket aman → status `PENDING` → publish ke **RabbitMQ/Kafka**.
4. Payment Service consume, simulasi bayar (jeda ~3 detik), publish `Payment.Success`.
5. Ticket Service potong kuota di DB utama; Notification Service log "tiket terkirim".

**Tantangan khas:** user pesan tapi tidak bayar dalam 5 menit → cron/worker otomatis batalkan pesanan dan kembalikan kuota.

### 2. Mini E-Commerce + Rekomendasi (fokus: event-driven & polyglot)

- Product Service (Go/Rust + PostgreSQL), Order Service (Go + MySQL), Recommendation Service (Python + MongoDB), Notification (Node.js).
- Event `Order.Created` → RabbitMQ → Recommendation Service hitung produk populer.

### 3. E-Wallet (fokus: konsistensi data, Saga Pattern)

- User/Account, Wallet, Ledger/Transaction Service.
- Masalah klasik: saldo A terpotong, transfer ke B gagal di tengah → butuh **Saga Pattern** dengan kompensasi/rollback otomatis.

### 4. Ojek Online Mini (fokus: real-time & geospatial)

- Passenger, Driver Tracker (GPS tiap 5 detik), Matching Engine, Trip Service.
- Teknologi: WebSockets + Redis Geo.

### 5. Video Transcoder (fokus: background worker)

- Upload Service → Transcoder Worker (FFmpeg, multi-resolusi) → Catalog Service.
- Object Storage: MinIO/S3.

## Ide Tanpa Pembayaran

| Proyek | Komunikasi | Senjata andalan | Tantangan |
|--------|-----------|-----------------|-----------|
| Matchmaking Game + Leaderboard | gRPC & WebSockets | Redis Sorted Sets | Real-time state, disconnect handling |
| Microblogging (Feed System) | REST + Message Broker | Caching Redis | Fan-out on write, tanpa SQL JOIN |
| Report Generator | Message Broker (Queue) | Worker & Storage | Tugas berat async agar tidak timeout |

## Peringkat Kerumitan

| Proyek | Level | Tantangan terbesar |
|--------|-------|--------------------|
| Report Generator | ⭐⭐ | Antrean kerja berat di background |
| Matchmaking Game | ⭐⭐⭐ | Koneksi real-time (WebSockets) konstan |
| Microblogging (Feed) | ⭐⭐⭐⭐ | Distribusi data massal (fan-out) tanpa JOIN — user 10.000 followers posting = 10.000 timeline update async |
| E-Commerce | ⭐⭐⭐⭐⭐ | Konsistensi data terdistribusi, Saga Pattern, rollback lintas service |
| Tiket Konser (war tiket) | ⭐⭐⭐⭐⭐ | 100.000 klik "Beli" di milidetik sama, stok 100 — distributed locking, anti-overselling |

## Kenapa Tiket, Padahal "Pasaran"?

Tech Lead tidak peduli **apa** aplikasinya, tapi **bagaimana** menyelesaikannya.

**Versi umum (tugas kuliah):** klik beli → `INSERT` langsung ke SQL. Diserbu 1.000 user → lock, overselling, error 500.

**Versi spesialis (yang dicari perusahaan):**
- **Redis Lua Scripting** — pemotongan kuota atomik, no overselling.
- **RabbitMQ queue** — DB utama tidak kena beban langsung.
- **Distributed Lock** — satu kursi tidak bisa dipesan dua orang di milidetik sama.

Judul boleh pasaran, arsitektur jangan.

## Alternatif Anti-Mainstream (kalau mau beda)

1. **Distributed Rate Limiter** — algoritma Token Bucket / Leaky Bucket di Redis, middleware di depan service lain.
2. **Centralized Log Collector + Alerting** — log dari semua service → broker → Elasticsearch/OpenSearch → alert error kritis.
3. **Real-Time Telemetry Dashboard** — streaming metrik (CPU, TPS), proses async, dashboard live tanpa refresh.

## Taktik Memulai (bertahap)

1. Buat Event/Ticket Service sampai bisa CRUD.
2. Buat Booking Service terpisah, hubungkan via REST (HTTP client biasa).
3. Bungkus keduanya ke Docker.
4. Install RabbitMQ via Docker Compose, ubah komunikasi jadi event-driven.
