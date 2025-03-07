BEGIN;

CREATE TABLE IF NOT EXISTS stock_rating (
    brokerage VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    company VARCHAR(100) NOT NULL,
    ticker VARCHAR(50) NOT NULL,
    rating_from VARCHAR(50) NOT NULL,
    rating_to VARCHAR(50) NOT NULL,
    target_from VARCHAR(50) NOT NULL,
    target_to VARCHAR(50) NOT NULL,
    time TIMESTAMP NOT NULL,
    target_price_change DECIMAL(10, 2) NOT NULL,
    PRIMARY KEY (brokerage, ticker),
    CONSTRAINT unique_record UNIQUE (brokerage, ticker, time)
);

CREATE INDEX idx_stock_rating_ticker_brokerage_time ON stock_rating (ticker, brokerage, time DESC)

COMMIT;