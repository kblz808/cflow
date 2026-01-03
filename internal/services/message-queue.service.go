package services

import (
	"cflow/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, event PaymentEvent) (ack bool, err error)

type PaymentEvent struct {
	PaymentID string `json:"payment_id"`
}

type MessageQueueService struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	url       string
	queueName string

	paymentService *PaymentService
}

func NewMessageQueueService(config *utils.RabbitMQConfig) (*MessageQueueService, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s", config.Username, config.Password, config.Host, config.Port)

	clientConfig := amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	}

	conn, err := amqp.DialConfig(url, clientConfig)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	if _, err := channel.QueueDeclare(
		config.QueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &MessageQueueService{
		conn:      conn,
		channel:   channel,
		url:       url,
		queueName: config.QueueName,
	}, nil
}

func (c *MessageQueueService) ConsumePayment(ctx context.Context, workerCount int, handler MessageHandler) error {
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

	for msg := range msgs {
		go func(m amqp.Delivery) {
			var event PaymentEvent
			if err := json.Unmarshal(m.Body, &event); err != nil {
				log.Printf("Invalid message format, rejecting: %v", err)
				m.Reject(false)
				return
			}

			ack, err := handler(ctx, event)
			if err != nil {
				log.Printf("Error processing payment %s: %v", event.PaymentID, err)
			}

			if ack {
				m.Ack(false)
			} else {
				m.Nack(false, true)
			}
		}(msg)

		select {
		case <-ctx.Done():
			log.Println("stopping rabbitmq consumer due to context cancellation")
			return nil
		default:
		}
	}

	return nil
}

func (c *MessageQueueService) PublishPaymentEvent(ctx context.Context, paymentID string) error {
	event := PaymentEvent{
		PaymentID: paymentID,
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

func (c *MessageQueueService) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *MessageQueueService) IsClosed() bool {
	return c.conn == nil || c.conn.IsClosed()
}
