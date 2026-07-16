package main

import (
	"context"
	"log"
	"time"

	"tix.at/internal/config"
	"tix.at/internal/stock"
	"tix.at/internal/store"
	"tix.at/internal/wait"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	var db *store.Store
	wait.For("postgres", func() (err error) {
		db, err = store.New(ctx, cfg.PostgresURL)
		return err
	})
	defer db.Close()

	redis := stock.New(cfg.RedisAddr)
	wait.For("redis", func() error { return redis.Ping(ctx) })
	defer redis.Close()

	log.Print("cancellation worker running")
	for {
		cancelExpired(ctx, db, redis)
		time.Sleep(cfg.CancelInterval)
	}
}

func cancelExpired(ctx context.Context, db *store.Store, redis *stock.Store) {
	bookings, err := db.ExpiredPending(ctx, 100)
	if err != nil {
		log.Printf("list expired failed: %v", err)
		return
	}
	for _, b := range bookings {
		ok, err := db.CancelExpired(ctx, b.ID)
		if err != nil {
			log.Printf("cancel %s failed: %v", b.ID, err)
			continue
		}
		if ok {
			if err := redis.Release(ctx, b.EventID); err != nil {
				log.Printf("release stock booking=%s failed: %v", b.ID, err)
			}
			log.Printf("booking cancelled booking=%s event=%s", b.ID, b.EventID)
		}
	}
}
