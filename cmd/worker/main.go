package main

import (
	"cflow/internal/models"
	"cflow/internal/repository"
	"cflow/internal/services"
	"cflow/internal/utils"
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

func main() {
	ctx := context.Background()

	dbConfig, err := utils.NewDBConfig()
	if err != nil {
		log.Fatalf("failed to load db config: %v", err)
	}

	mqConfig, err := utils.NewMQConfig()
	if err != nil {
		log.Fatalf("failed to load mq config: %v", err)
	}

	db, err := repository.NewDB(ctx, dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	mqClient, err := services.NewMessageQueueService(mqConfig)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer mqClient.Close()

	repo := repository.NewPaymentRepository(db)
	svc := services.NewPaymentService(repo, mqClient)

	handler := func(ctx context.Context, event services.PaymentEvent) (bool, error) {
		log.Printf("Processing payment: %s", event.PaymentID)

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		targetStatus := models.StatusSuccess
		if rand.Float32() >= 0.5 {
			targetStatus = models.StatusFailed
		}

		log.Printf("Attempting to mark payment %s as %s", event.PaymentID, targetStatus)
		if err := svc.ChangePaymentStatus(ctx, event.PaymentID, targetStatus); err != nil {
			if errors.Is(err, repository.ErrPaymentAlreadyProcessed) {
				log.Printf("Payment %s is already processed, skipping", event.PaymentID)
				return true, nil
			}
			if errors.Is(err, repository.ErrPaymentNotFound) {
				log.Printf("Payment %s not found, skipping", event.PaymentID)
				return true, nil
			}
			log.Printf("Failed to process payment %s: %v", event.PaymentID, err)
			return false, err
		}

		log.Printf("Successfully marked payment %s as %s", event.PaymentID, targetStatus)
		return true, nil
	}

	workerCount := 10
	if err := mqClient.ConsumePayment(ctx, workerCount, handler); err != nil {
		log.Fatalf("consumer error: %v", err)
	}
}
