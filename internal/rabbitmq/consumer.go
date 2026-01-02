package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type MessageHandler func(ctx context.Context, event PaymentEvent) (ack bool, err error)

func (c *Client) ConsumePayment(ctx context.Context, workerCount int, handler MessageHandler) error {
	if err := c.channel.Qos(
		workerCount,
		0,
		false,
	); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("RabbitMQ consumer started with %d workers", workerCount)

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping rabbitmq consumer due to context cancellation")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("rabbitmq consumer channel closed")
				return nil
			}

			var event PaymentEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Invalid message format, rejecting: %v", err)
				msg.Reject(false)
				continue
			}

			ack, err := handler(ctx, event)
			if err != nil {
				log.Printf("Error processing payment %s: %v", event.PaymentID, err)
			}

			if ack {
				msg.Ack(false)
			} else {
				msg.Nack(false, true)
			}

		}
	}
}
