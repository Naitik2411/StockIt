-- Write your migrate up statements here
CREATE TABLE stocks(
    ticker VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT ' ',
    current_price NUMERIC(20,6) NOT NULL DEFAULT 0,
    price_change_pct NUMERIC(10, 4) NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE IF EXISTS stocks;