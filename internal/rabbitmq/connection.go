package rabbitmq

import (
	"cflow/internal/utils"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	url       string
	queueName string
}

func NewClient(config *utils.RabbitMQConfig) (*Client, error) {
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

	return &Client{
		conn:      conn,
		channel:   channel,
		url:       url,
		queueName: config.QueueName,
	}, nil
}

func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *Client) IsClosed() bool {
	return c.conn == nil || c.conn.IsClosed()
}
