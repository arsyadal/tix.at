package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"tix.at/internal/config"
	"tix.at/internal/events"
	"tix.at/internal/rabbit"
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

	var mq *rabbit.Client
	wait.For("rabbitmq", func() (err error) {
		mq, err = rabbit.New(cfg.RabbitURL, events.BookingCreatedQueue, events.PaymentSuccessQueue)
		return err
	})
	defer mq.Close()

	deliveries, err := mq.Consume(events.BookingCreatedQueue, "payment-worker")
	must(err)
	log.Print("payment worker running")

	for d := range deliveries {
		var msg events.BookingCreated
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			_ = d.Nack(false, false)
			continue
		}

		time.Sleep(cfg.PaymentDelay)
		paid, err := db.MarkPaid(ctx, msg.BookingID, msg.EventID)
		if err != nil {
			log.Printf("mark paid failed: %v", err)
			_ = d.Nack(false, true)
			continue
		}
		if !paid {
			_ = d.Ack(false)
			continue
		}
		if err := mq.PublishJSON(ctx, events.PaymentSuccessQueue, events.PaymentSuccess(msg)); err != nil {
			log.Printf("publish payment success failed: %v", err)
			_ = d.Nack(false, true)
			continue
		}
		_ = d.Ack(false)
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
