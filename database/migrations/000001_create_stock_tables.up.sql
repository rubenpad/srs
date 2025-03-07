BEGIN;

CREATE TABLE IF NOT EXISTS stock (
    ticket CHAR(50) PRIMARY KEY,
    company CHAR(50) NOT NULL,
    score DECIMAL(10, 5) NOT NULL
);

CREATE TABLE IF NOT EXISTS stock_rating (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brokerage CHAR(50) NOT NULL,
    action CHAR(50) NOT NULL,
    company CHAR(50) NOT NULL,
    ticker CHAR(50) NOT NULL,
    rating_from CHAR(50) NOT NULL,
    rating_to CHAR(50) NOT NULL,
    target_from DECIMAL(10, 5) NOT NULL,
    target_to DECIMAL(10, 5) NOT NULL,
    time TIMESTAMP NOT NULL,
    target_price_change DECIMAL(10, 5) NOT NULL
);

COMMIT;