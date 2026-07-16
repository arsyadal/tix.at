package config

import (
	"os"
	"time"
)

type Config struct {
	Addr           string
	PostgresURL    string
	RedisAddr      string
	RabbitURL      string
	PaymentDelay   time.Duration
	BookingTTL     time.Duration
	CancelInterval time.Duration
}

func Load() Config {
	return Config{
		Addr:           env("ADDR", ":8080"),
		PostgresURL:    env("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/tix?sslmode=disable"),
		RedisAddr:      env("REDIS_ADDR", "localhost:6379"),
		RabbitURL:      env("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		PaymentDelay:   durationEnv("PAYMENT_DELAY", 3*time.Second),
		BookingTTL:     durationEnv("BOOKING_TTL", 5*time.Minute),
		CancelInterval: durationEnv("CANCEL_INTERVAL", 10*time.Second),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
