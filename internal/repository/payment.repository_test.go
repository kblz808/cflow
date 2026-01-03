package repository

import (
	"cflow/internal/models"
	"cflow/internal/utils"
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRepo *PaymentRepository
var testDB *DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	config, err := utils.NewDBConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	db, err := NewDB(ctx, config)
	if err != nil {
		panic("failed to connect to db: " + err.Error())
	}
	testDB = db
	testRepo = NewPaymentRepository(db)

	if err := db.Migrate(); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	code := m.Run()

	db.Close()
	os.Exit(code)
}

func TestPaymentRepository_CreatePayment(t *testing.T) {
	ctx := context.Background()
	ref := uuid.New().String()

	payment := &models.Payment{
		Amount:    100.0,
		Currency:  models.CurrencyUSD,
		Reference: ref,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		err := testRepo.CreatePayment(ctx, payment)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, payment.ID)
	})

	t.Run("DuplicateReference", func(t *testing.T) {
		err := testRepo.CreatePayment(ctx, payment)
		assert.ErrorIs(t, err, ErrDuplicateReference)
	})
}

func TestPaymentRepository_GetPaymentByID(t *testing.T) {
	ctx := context.Background()
	ref := uuid.New().String()

	payment := &models.Payment{
		Amount:    200.0,
		Currency:  models.CurrencyETB,
		Reference: ref,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := testRepo.CreatePayment(ctx, payment)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		found, err := testRepo.GetPaymentByID(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Equal(t, payment.Amount, found.Amount)
		assert.Equal(t, payment.Reference, found.Reference)
		assert.Equal(t, payment.Status, found.Status)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := testRepo.GetPaymentByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrPaymentNotFound)
	})
}

func TestPaymentRepository_UpdatePaymentStatus(t *testing.T) {
	ctx := context.Background()
	ref := uuid.New().String()

	payment := &models.Payment{
		Amount:    300.0,
		Currency:  models.CurrencyUSD,
		Reference: ref,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := testRepo.CreatePayment(ctx, payment)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := testRepo.UpdatePaymentStatus(ctx, payment.ID, models.StatusSuccess)
		assert.NoError(t, err)

		found, _ := testRepo.GetPaymentByID(ctx, payment.ID)
		assert.Equal(t, models.StatusSuccess, found.Status)
	})

	t.Run("AlreadyProcessed", func(t *testing.T) {
		err := testRepo.UpdatePaymentStatus(ctx, payment.ID, models.StatusFailed)
		assert.ErrorIs(t, err, ErrPaymentAlreadyProcessed)
	})

	t.Run("NotFound", func(t *testing.T) {
		err := testRepo.UpdatePaymentStatus(ctx, uuid.New(), models.StatusSuccess)
		assert.ErrorIs(t, err, ErrPaymentNotFound)
	})
}
