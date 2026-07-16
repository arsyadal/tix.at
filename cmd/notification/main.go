package main

import (
	"encoding/json"
	"log"

	"tix.at/internal/config"
	"tix.at/internal/events"
	"tix.at/internal/rabbit"
	"tix.at/internal/wait"
)

func main() {
	cfg := config.Load()
	var mq *rabbit.Client
	wait.For("rabbitmq", func() (err error) {
		mq, err = rabbit.New(cfg.RabbitURL, events.PaymentSuccessQueue)
		return err
	})
	defer mq.Close()

	deliveries, err := mq.Consume(events.PaymentSuccessQueue, "notification")
	must(err)
	log.Print("notification worker running")

	for d := range deliveries {
		var msg events.PaymentSuccess
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			_ = d.Nack(false, false)
			continue
		}
		log.Printf("ticket sent booking=%s user=%s event=%s", msg.BookingID, msg.UserID, msg.EventID)
		_ = d.Ack(false)
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
