BEGIN;

CREATE TABLE IF NOT EXISTS stock_rating (
    ticker VARCHAR(50) NOT NULL,
    brokerage VARCHAR(50) NOT NULL,
    time TIMESTAMP NOT NULL,
    action VARCHAR(50) NOT NULL,
    company VARCHAR(100) NOT NULL,
    rating_from VARCHAR(50) NOT NULL,
    rating_to VARCHAR(50) NOT NULL,
    target_from VARCHAR(50) NOT NULL,
    target_to VARCHAR(50) NOT NULL,
    target_price_change DECIMAL(10, 2) NOT NULL,
    score DECIMAL(10, 2) NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (ticker, brokerage, time DESC)
);

COMMIT;