package repository

import (
	"cflow/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
	ErrDuplicateReference      = errors.New("duplicate reference")
)

type PaymentRepository struct {
	db *DB
}

func NewPaymentRepository(db *DB) *PaymentRepository {
	return &PaymentRepository{db}
}

func (repo *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := repo.db.QueryBuilder.Insert("payments").
		Columns("amount", "currency", "reference", "status", "created_at", "updated_at").
		Values(payment.Amount, payment.Currency, payment.Reference, payment.Status, payment.CreatedAt, payment.UpdatedAt).
		Suffix("RETURNING id, created_at, updated_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&payment.ID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if repo.db.ErrorCode(err) == "23505" {
			return ErrDuplicateReference
		}
		return err
	}

	return tx.Commit(ctx)
}

func (repo *PaymentRepository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := repo.db.QueryBuilder.Select("id, amount, currency, reference, status, created_at").
		From("payments").
		Where("id = ?", id)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var payment models.Payment
	err = tx.QueryRow(ctx, sql, args...).Scan(
		&payment.ID,
		&payment.Amount,
		&payment.Currency,
		&payment.Reference,
		&payment.Status,
		&payment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return &payment, tx.Commit(ctx)
}

func (repo *PaymentRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status models.Status) error {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentStatus models.Status
	statusQuery := repo.db.QueryBuilder.Select("status").
		From("payments").
		Where("id = ?", id).
		Suffix("FOR UPDATE")

	sql, args, err := statusQuery.ToSql()
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return err
	}

	if currentStatus != models.StatusPending {
		return ErrPaymentAlreadyProcessed
	}

	updateQuery := repo.db.QueryBuilder.Update("payments").
		Set("status", status).
		Set("updated_at", time.Now()).
		Where("id = ?", id)

	sql, args, err = updateQuery.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
