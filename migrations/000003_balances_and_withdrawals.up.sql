CREATE TABLE IF NOT EXISTS balances
(
    id         SERIAL PRIMARY KEY,
    user_id    INT            NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    current    DECIMAL(15, 2) NOT NULL  DEFAULT 0.0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS withdrawals
(
    id               SERIAL PRIMARY KEY,
    user_id          INT            NOT NULL,
    balance_id       INT            NOT NULL,
    order_numeral_id TEXT           NOT NULL,
    amount           DECIMAL(15, 2) NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (balance_id) REFERENCES balances (id) ON DELETE CASCADE
);

CREATE INDEX idx_balances_user_id ON balances (user_id);
CREATE INDEX idx_withdrawals_balance_id ON withdrawals (balance_id);
CREATE INDEX idx_withdrawals_order_id ON withdrawals (order_numeral_id) WHERE order_numeral_id IS NOT NULL;
