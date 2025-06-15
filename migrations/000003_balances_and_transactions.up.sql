CREATE TABLE IF NOT EXISTS balances
(
    id         SERIAL PRIMARY KEY,
    user_id    INT            NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    current    DECIMAL(15, 2) NOT NULL  DEFAULT 0.0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions
(
    id         SERIAL PRIMARY KEY,
    user_id    INT            NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    balance_id INT            NOT NULL,
    FOREIGN KEY (balance_id) REFERENCES balances (id) ON DELETE CASCADE,
    order_numeral_id      TEXT           NOT NULL,
    FOREIGN KEY (order_numeral_id) REFERENCES orders (numeral_id) ON DELETE SET NULL,
    amount     DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_balances_user_id ON balances (user_id);
CREATE INDEX idx_transactions_balance_id ON transactions (balance_id);
CREATE INDEX idx_transactions_order_id ON transactions (order_numeral_id) WHERE order_numeral_id IS NOT NULL;
