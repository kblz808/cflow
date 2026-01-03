package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Connection string
	Host       string
	Port       string
	Username   string
	Password   string
	Name       string
	MaxConns   int
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		c.Connection, c.Username, c.Password, c.Host, c.Port, c.Name,
	)
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string

	QueueName   string
	WorkerCount int
}

func NewDBConfig() (*DBConfig, error) {
	_ = godotenv.Load()

	db := DBConfig{
		Connection: os.Getenv("DB_CONNECTION"),
		Host:       os.Getenv("DB_HOST"),
		Port:       os.Getenv("DB_PORT"),
		Username:   os.Getenv("DB_USER"),
		Password:   os.Getenv("DB_PASSWORD"),
		Name:       os.Getenv("DB_NAME"),
	}

	maxConnsStr := os.Getenv("DB_MAX_CONNS")
	maxConns, err := strconv.Atoi(maxConnsStr)
	if err != nil || maxConns <= 0 {
		maxConns = 50
	}
	db.MaxConns = maxConns

	return &db, nil
}

func NewMQConfig() (*RabbitMQConfig, error) {
	_ = godotenv.Load()

	mq := RabbitMQConfig{
		Host:      os.Getenv("MQ_HOST"),
		Port:      os.Getenv("MQ_PORT"),
		Username:  os.Getenv("MQ_USER"),
		Password:  os.Getenv("MQ_PASSWORD"),
		QueueName: os.Getenv("MQ_QUEUE"),
	}

	workerCountStr := os.Getenv("MQ_WORKER_COUNT")
	workerCount, err := strconv.Atoi(workerCountStr)
	if err != nil || workerCount <= 0 {
		workerCount = 10
	}
	mq.WorkerCount = workerCount

	return &mq, nil
}
