package main

import (
	"cflow/internal/repository"
	"cflow/internal/services"
	"cflow/internal/utils"
	"context"
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

		status, err := svc.GetPaymentStatus(ctx, event.PaymentID)
		if err != nil {
			log.Printf("Failed to get payment status: %v", err)
			return false, err
		}

		if status == "SUCCESS" {
			log.Printf("Payment %s is already SUCCESS", event.PaymentID)
			return true, nil
		}

		if status == "FAILED" {
			log.Printf("Payment %s is already FAILED", event.PaymentID)
			return true, nil
		}

		if rand.Float32() < 0.5 {
			log.Printf("Marking payment %s as SUCCESS", event.PaymentID)
			if err := svc.MarkPaymentAsSuccess(ctx, event.PaymentID); err != nil {
				log.Printf("Failed to mark payment %s as success: %v", event.PaymentID, err)
				return false, err
			}
		} else {
			log.Printf("Marking payment %s as FAILED", event.PaymentID)
			if err := svc.MarkPaymentAsFailed(ctx, event.PaymentID); err != nil {
				log.Printf("Failed to mark payment %s as failed: %v", event.PaymentID, err)
				return false, err
			}
		}

		return true, nil
	}

	workerCount := 10
	if err := mqClient.ConsumePayment(ctx, workerCount, handler); err != nil {
		log.Fatalf("consumer error: %v", err)
	}
}
