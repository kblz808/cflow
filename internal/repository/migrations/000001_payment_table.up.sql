CREATE TABLE IF NOT EXISTS payments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount     NUMERIC(20, 8) NOT NULL CHECK (amount > 0),
    currency   VARCHAR(3)    NOT NULL CHECK (currency IN ('ETB', 'USD')),
    reference  VARCHAR(255)  NOT NULL,
    status     VARCHAR(20)   NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED')),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    UNIQUE (reference)
);

CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

CREATE INDEX IF NOT EXISTS idx_payments_pending ON payments(id) WHERE status = 'PENDING';
