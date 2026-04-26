CREATE TABLE orders (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        number TEXT NOT NULL UNIQUE,
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        status TEXT NOT NULL DEFAULT 'NEW',
                        accrual DECIMAL(12,2),
                        uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);