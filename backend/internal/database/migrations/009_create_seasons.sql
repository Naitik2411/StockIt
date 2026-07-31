-- Write your migrate up statements here
CREATE TABLE seasons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    season_number INT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'completed', 'pending')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(league_id, season_number)
);

CREATE INDEX idx_seasons_league_status ON seasons(league_id, status);
CREATE INDEX idx_seasons_end_date ON seasons(end_date) WHERE status = 'active';

---- create above / drop below ----

DROP TABLE IF EXISTS seasons;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
