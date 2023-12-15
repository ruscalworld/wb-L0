package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"wb-l0/internal/config"
	"wb-l0/internal/order"

	"github.com/nats-io/nats.go"
)

type Consumer struct {
	conn    *nats.Conn
	subject string

	orderRepository order.Repository
}

func NewConsumer(cfg config.NatsConnection, orderRepository order.Repository) (*Consumer, error) {
	conn, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to nats: %s", err)
	}

	return &Consumer{
		conn:    conn,
		subject: cfg.Subject,

		orderRepository: orderRepository,
	}, nil
}

func (c *Consumer) Subscribe(ctx context.Context) error {
	_, err := c.conn.Subscribe(c.subject, c.wrappedMessageHandler(ctx)) // TODO: start from sequence id

	if err != nil {
		return fmt.Errorf("error subscribing to subject \"%s\": %s", c.subject, err)
	}

	log.Printf("subscribed to subject \"%s\"\n", c.subject)
	return nil
}

func (c *Consumer) wrappedMessageHandler(ctx context.Context) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		log.Printf("received message with length of %d bytes\n", len(msg.Data))
		err := c.handleMessage(ctx, msg)
		if err != nil {
			log.Printf("discarding message due to error: %s", err)
			return
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg *nats.Msg) error {
	var o order.Order
	err := json.Unmarshal(msg.Data, &o)
	if err != nil {
		return fmt.Errorf("error unmarshalling message: %s\n", err)
	}

	err = o.Validate()
	if err != nil {
		log.Printf("message has failed validation: %s\n", err)
		return err
	}

	err = c.orderRepository.CreateOrder(ctx, &o)
	if err != nil {
		log.Printf("error saving message: %s\n", err)
		return err
	}

	log.Println("saved new order with UID", o.OrderUID)
	return nil
}
