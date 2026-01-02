package repository

import (
	"cflow/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentRepository struct {
	db *DB
}

func NewPaymentRepository(db *DB) PaymentRepository {
	return PaymentRepository{db}
}

func (repo *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	query := repo.db.QueryBuilder.Insert("payments").
		Columns("amount", "currency", "reference", "status", "created_at", "updated_at").
		Values(payment.Amount, payment.Currency, payment.Reference, payment.Status, payment.CreatedAt, payment.UpdatedAt).
		Suffix("RETURNING id, created_at, updated_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	err = repo.db.QueryRow(ctx, sql, args...).Scan(
		&payment.ID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (repo *PaymentRepository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	query := repo.db.QueryBuilder.Select("amount, currency, reference, status, created_at").
		From("payments").
		Where("id = ?", id)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var payment models.Payment
	err = repo.db.QueryRow(ctx, sql, args...).Scan(
		&payment.Amount,
		&payment.Currency,
		&payment.Reference,
		&payment.Status,
		&payment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (repo *PaymentRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status models.Status) error {
	query := repo.db.QueryBuilder.Update("payments").
		Set("status", status).
		Set("updated_at", time.Now()).
		Where("id = ?", id)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = repo.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
