-- Write your migrate up statements here

CREATE TABLE season_snapshots(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    final_value NUMERIC(20, 6) NOT NULL,
    return_pct NUMERIC(10,4) NOT NULL,
    rank INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(season_id, user_id)
);

CREATE INDEX idx_snapshots_season ON season_snapshots(season_id, rank);
CREATE INDEX idx_snapshots_user ON season_snapshots(user_id);

---- create above / drop below ----
DROP TABLE IF EXISTS season_snapshots;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
