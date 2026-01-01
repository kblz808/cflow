package repository

type PaymentRepository struct {
	db *DB
}

func NewPaymentRepository(db *DB) PaymentRepository {
	return PaymentRepository{db}
}
