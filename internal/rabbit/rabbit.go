package rabbit

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New(url string, queues ...string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	for _, q := range queues {
		if _, err := ch.QueueDeclare(q, true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}
	}
	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Close() {
	_ = c.ch.Close()
	_ = c.conn.Close()
}

func (c *Client) PublishJSON(ctx context.Context, queue string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (c *Client) Consume(queue, consumer string) (<-chan amqp.Delivery, error) {
	if err := c.ch.Qos(10, 0, false); err != nil {
		return nil, err
	}
	return c.ch.Consume(queue, consumer, false, false, false, false, nil)
}
