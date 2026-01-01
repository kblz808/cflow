package utils

import (
	"github.com/joho/godotenv"
	"os"
)

type DBConfig struct {
	Connection string
	Host       string
	Port       string
	Username   string
	Password   string
	Name       string
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

func NewDBConfig() (*DBConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	db := DBConfig{
		Connection: os.Getenv("DB_CONNECTION"),
		Host:       os.Getenv("DB_HOST"),
		Port:       os.Getenv("DB_PORT"),
		Username:   os.Getenv("DB_USER"),
		Password:   os.Getenv("DB_PASSWORD"),
		Name:       os.Getenv("DB_NAME"),
	}

	return &db, nil
}

func NewMQConfig() (*RabbitMQConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	mq := RabbitMQConfig{
		Host:     os.Getenv("MQ_HOST"),
		Port:     os.Getenv("MQ_PORT"),
		Username: os.Getenv("MQ_USER"),
		Password: os.Getenv("MQ_PASSWORD"),
	}

	return &mq, nil
}
