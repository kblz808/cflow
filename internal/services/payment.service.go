package services

import (
	"cflow/internal/models"
	"cflow/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	repo *repository.PaymentRepository
	mq   *MessageQueueService
}

func NewPaymentService(repo *repository.PaymentRepository, mq *MessageQueueService) *PaymentService {
	return &PaymentService{
		repo: repo,
		mq:   mq,
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

	if err := s.repo.CreatePayment(ctx, &payment); err != nil {
		return err
	}

	return s.mq.PublishPaymentEvent(ctx, payment.ID.String())
}

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*models.GetPaymentResponse, error) {
	payment, err := s.repo.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.GetPaymentResponse{
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Reference: payment.Reference,
		Status:    payment.Status,
		CreatedAt: payment.CreatedAt,
	}, nil
}

func (s *PaymentService) MarkPaymentAsSuccess(ctx context.Context, id string) error {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.Status != models.StatusPending {
		return fmt.Errorf("payment is not in pending state")
	}

	return s.repo.UpdatePaymentStatus(ctx, paymentID, models.StatusSuccess)
}

func (s *PaymentService) MarkPaymentAsFailed(ctx context.Context, id string) error {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.Status != models.StatusPending {
		return fmt.Errorf("payment is not in pending state")
	}

	return s.repo.UpdatePaymentStatus(ctx, paymentID, models.StatusFailed)
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, id string) (models.Status, error) {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid id: %w", err)
	}

	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return "", fmt.Errorf("failed to get payment: %w", err)
	}

	return payment.Status, nil
}
