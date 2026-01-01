package services

import (
	"cflow/internal/models"
	"cflow/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		repo: repo,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, paymentRequest *models.CreatePaymentRequest) error {
	payment := models.Payment{
		ID:        uuid.New(),
		Amount:    paymentRequest.Amount,
		Currency:  models.Currency(paymentRequest.Currency),
		Reference: paymentRequest.Reference,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.CreatePayment(ctx, &payment)
}

func (s *PaymentService) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	return s.repo.GetPaymentByID(ctx, id)
}
