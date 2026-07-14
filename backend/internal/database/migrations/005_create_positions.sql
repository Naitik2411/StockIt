-- Write your migrate up statements here
CREATE TABLE positions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    ticker VARCHAR(20) NOT NULL,
    shares NUMERIC(20, 6) NOT NULL CHECK(shares > 0),
    avg_buy_price NUMERIC(20, 6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(portfolio_id, ticker)
);

CREATE INDEX idx_positions_portfolio ON positions(portfolio_id);
---- create above / drop below ----
DROP TABLE IF EXISTS positions;     
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
