CREATE TABLE withdrawals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number TEXT NOT NULL,
    sum DECIMAL(12,2) NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, order_number)
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);