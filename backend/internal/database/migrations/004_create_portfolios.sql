-- Write your migrate up statements here
CREATE TABLE portfolios(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cash_balance NUMERIC(20, 6) NOT NULL DEFAULT 100000.000000,
    season_id VARCHAR(20) NOT NULL DEFAULT 'global-2025',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, season_id)
);

CREATE INDEX idx_portfolios_user_id ON portfolios(user_id);
---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE IF EXISTS portfolios;