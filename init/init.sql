CREATE TABLE IF NOT EXISTS stocks (
    name VARCHAR(255) PRIMARY KEY,
    quantity INT NOT NULL CHECK (quantity >= 0)
);

CREATE TABLE IF NOT EXISTS wallet_stocks (
    wallet_id VARCHAR(255) NOT NULL,
    stock_name VARCHAR(255) REFERENCES stocks(name),
    quantity INT NOT NULL CHECK (quantity >= 0),
    PRIMARY KEY (wallet_id, stock_name)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    type VARCHAR(10) NOT NULL, -- 'buy' or 'sell'
    wallet_id VARCHAR(255) NOT NULL,
    stock_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);