package models

import (
	"time"

	"github.com/google/uuid"
)

type Currency string
type Status string

const (
	CurrencyETB Currency = "ETB"
	CurrencyUSD Currency = "USD"

	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)

type Payment struct {
	ID        uuid.UUID `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  Currency  `json:"currency"`
	Reference string    `json:"reference"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreatePaymentRequest struct {
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Currency  string  `json:"currency" validate:"required,oneof=ETB USD"`
	Reference string  `json:"reference" validate:"required"`
}

type GetPaymentResponse struct {
	Amount    float64   `json:"amount"`
	Currency  Currency  `json:"currency"`
	Reference string    `json:"reference"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
