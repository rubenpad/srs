BEGIN;

CREATE TABLE IF NOT EXISTS stock (
    ticker VARCHAR(50) PRIMARY KEY,
    company VARCHAR(100) NOT NULL,
    score DECIMAL(10, 2) NOT NULL
);

CREATE TABLE IF NOT EXISTS stock_rating (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brokerage VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    company VARCHAR(100) NOT NULL,
    ticker VARCHAR(50) NOT NULL,
    rating_from VARCHAR(50) NOT NULL,
    rating_to VARCHAR(50) NOT NULL,
    target_from VARCHAR(50) NOT NULL,
    target_to VARCHAR(50) NOT NULL,
    time TIMESTAMP NOT NULL,
    target_price_change DECIMAL(10, 2) NOT NULL
);

COMMIT;