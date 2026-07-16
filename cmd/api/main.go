package main

import (
	"context"
	"log"
	"net/http"

	"tix.at/internal/config"
	"tix.at/internal/events"
	"tix.at/internal/httpapi"
	"tix.at/internal/rabbit"
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

	var mq *rabbit.Client
	wait.For("rabbitmq", func() (err error) {
		mq, err = rabbit.New(cfg.RabbitURL, events.BookingCreatedQueue)
		return err
	})
	defer mq.Close()

	log.Printf("api listening on %s", cfg.Addr)
	must(http.ListenAndServe(cfg.Addr, httpapi.New(db, redis, mq, cfg.BookingTTL).Handler()))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
