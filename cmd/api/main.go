package main

import (
	"cflow/internal/handlers"
	"cflow/internal/repository"
	"cflow/internal/services"
	"cflow/internal/utils"
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	config, err := utils.NewDBConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := repository.NewDB(ctx, config)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	mqConfig, err := utils.NewMQConfig()
	if err != nil {
		log.Fatalf("failed to load mq config: %v", err)
	}

	mqClient, err := services.NewMessageQueueService(mqConfig)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer mqClient.Close()

	repo := repository.NewPaymentRepository(db)
	svc := services.NewPaymentService(repo, mqClient)
	handler := handlers.NewPaymentHandler(svc)

	router := handlers.NewRouter(handler)

	log.Fatal(router.Start(":8000"))
}
