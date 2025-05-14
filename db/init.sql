CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    base_currency VARCHAR(3),
    target_currency VARCHAR(3),
    rate DECIMAL,
    scraped_at TIMESTAMP
);
