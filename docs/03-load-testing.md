# Load Testing — Simulasi Ribuan Pengguna Tanpa Pengguna Asli

Tidak butuh user asli. Engineer di dunia nyata menguji dengan pengguna palsu (load/stress testing) dari laptop sendiri.

## 1. Tools Load Testing

| Tool | Bahasa skrip | Kelebihan |
|------|-------------|-----------|
| **k6** (rekomendasi) | JavaScript/TypeScript (engine Go) | Cepat, hemat memori, developer-friendly |
| **Locust** | Python | Grafik performa real-time via web UI |
| **Apache Bench (ab)** | CLI satu baris | Paling simpel, bawaan terminal |

Contoh perintah k6: "tembak `POST /booking` 1.000 kali dalam 5 detik."

## 2. Batasi Resource Container Docker

Laptop modern terlalu kuat — aplikasi terasa aman padahal belum teruji. Batasi di `docker-compose.yml` agar terasa seperti server murah:

```yaml
services:
  order-service:
    deploy:
      resources:
        limits:
          cpus: "0.25"
          memory: 128M
```

Tembak dengan k6 500 user palsu → server mulai megap-megap: `502 Bad Gateway`, timeout, atau container crash. Di situ belajarnya.

## 3. Inject Latency (Simulasi Jaringan Lambat)

Localhost terlalu cepat untuk mensimulasikan jaringan nyata.

- Cara manual: `time.Sleep(2 * time.Second)` di Payment Service — pura-pura bank lemot.
- Efek domino terlihat: Booking Service ikut melambat, memori menumpuk karena menunggu.

## 4. Skenario Latihan Bertahap

1. **Baseline** — aplikasi booking mini tanpa pembatasan. Tembak 200 users/sec. Cek database aman atau tidak.
2. **Chaos** — sengaja matikan database Payment Service. Coba pesan tiket. Apakah Booking Service ikut crash atau memberi error rapi?
3. **Optimasi** — pasang Redis cache di depan DB, tembak 5x lipat lebih banyak. Bandingkan grafik response time before/after.

Hasil before/after ini yang masuk README sebagai bukti (lihat `04-portofolio.md`).
