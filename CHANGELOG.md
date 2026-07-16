# tix.at — Changelog & Roadmap (Pindah Laptop)

Status terakhir sebelum migrasi ke laptop Mac.

## Status Saat Ini

- **Dokumentasi Lengkap**: Semua konsep arsitektur, ide proyek, strategi load testing, dan panduan portfolio telah disimpan di folder `docs/`.
- **Environment**: Go sudah terinstal secara lokal di laptop asal, tetapi pengerjaan implementasi kode ditangguhkan (hold) agar bisa dimulai langsung secara bersih di laptop Mac.

---

## Rencana Implementasi di Laptop Mac

Ketika Anda melanjutkan proyek ini di Mac, berikut adalah langkah-langkah yang perlu dijalankan:

### Langkah 0: Persiapan Environment di Mac
1. Pastikan **Go (versi 1.22+)** terinstal (`brew install go`).
2. Pastikan **Docker** & **Docker Desktop** terinstal dan berjalan (`brew install --cask docker`).
3. Clone repo ini: `git clone https://github.com/arsyadal/tix.at.git`

---

### Langkah 1: Implementasi Fase 1 (Ticket Service + Postgres)

1. Buat folder dan inisialisasi module:
   ```bash
   mkdir -p services/ticket
   cd services/ticket
   go mod init tix.at/services/ticket
   # Install pgx driver
   go get github.com/jackc/pgx/v5
   ```
2. Buat file `services/ticket/main.go` menggunakan standard library (`net/http` dengan route patterns Go 1.22+):
   - `GET /events`
   - `GET /events/{id}`
   - `POST /events` (dengan validasi input nama, kuota > 0, price >= 0)
3. Tambahkan auto-create table saat startup:
   ```sql
   CREATE TABLE IF NOT EXISTS events (
       id SERIAL PRIMARY KEY,
       name VARCHAR(255) NOT NULL,
       venue VARCHAR(255) NOT NULL,
       event_date TIMESTAMP NOT NULL,
       quota INT NOT NULL,
       price DECIMAL(10, 2) NOT NULL
   );
   ```
4. Buat `services/ticket/Dockerfile`.
5. Buat `docker-compose.yml` di root directory yang menjalankan `ticket-service` dan container `postgres` terpisah.
6. Verifikasi menggunakan curl/Postman dan jalankan manual.

---

### Langkah 2: Booking Service (Fase 2)
1. Buat `services/booking/` dan hubungkan dengan Ticket Service via HTTP REST client.
2. Simpan status booking `PENDING` di DB PostgreSQL booking tersendiri (Database per Service).

### Langkah 3: Redis Stock Counter (Fase 3)
1. Tambahkan Redis ke `docker-compose.yml`.
2. Sync kuota ke Redis saat event dibuat. Cek sisa kuota dan kurangi kuota secara atomik menggunakan Redis Lua Scripting sebelum membuat booking.

### Langkah 4: RabbitMQ & Worker (Fase 4)
1. Tambahkan RabbitMQ ke `docker-compose.yml`.
2. Gunakan pattern event-driven untuk transaksi pembayaran asinkron dan notification worker.

### Langkah 5: Load Testing (Fase 5)
1. Batasi resource container di docker-compose (0.25 CPU / 128MB).
2. Buat skrip k6 di folder `loadtest/` dan jalankan stress testing.
