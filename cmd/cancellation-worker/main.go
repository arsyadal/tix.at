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
	var lastReconcile time.Time
	for {
		cancelExpired(ctx, db, redis)
		if time.Since(lastReconcile) >= 30*time.Second {
			reconcileStocks(ctx, db, redis)
			lastReconcile = time.Now()
		}
		time.Sleep(cfg.CancelInterval)
	}
}

func reconcileStocks(ctx context.Context, db *store.Store, redis *stock.Store) {
	stocks, err := db.GetCalculatedStocks(ctx)
	if err != nil {
		log.Printf("reconcile list stocks failed: %v", err)
		return
	}
	for _, s := range stocks {
		curr, err := redis.Get(ctx, s.EventID)
		if err != nil {
			log.Printf("reconcile get redis stock event=%s failed: %v", s.EventID, err)
			continue
		}
		if curr != s.CalculatedStock {
			log.Printf("reconcile detected drift event=%s db=%d redis=%d, correcting redis", s.EventID, s.CalculatedStock, curr)
			if err := redis.Set(ctx, s.EventID, s.CalculatedStock); err != nil {
				log.Printf("reconcile set redis stock event=%s failed: %v", s.EventID, err)
			}
		}
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
