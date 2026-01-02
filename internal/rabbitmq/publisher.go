package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

func (c *Client) PublishPaymentEvent(ctx context.Context, paymentID string, status string) error {
	event := PaymentEvent{
		PaymentID: paymentID,
		Status:    status,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",
		c.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish payment event: %w", err)
	}

	return nil
}
