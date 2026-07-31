-- Write your migrate up statements here
ALTER TABLE portfolios
    ADD COLUMN league_id UUID REFERENCES leagues(id) NULL,
    ADD COLUMN season_ref_id UUID REFERENCES seasons(id) NULL;
-- Old UNIQUE(user_id, season_id) breaks once a user has global + league portfolios.
ALTER TABLE portfolios DROP CONSTRAINT IF EXISTS portfolios_user_id_season_id_key;
CREATE UNIQUE INDEX idx_portfolios_global_user
    ON portfolios(user_id)
    WHERE league_id IS NULL;
CREATE UNIQUE INDEX idx_portfolios_league_season_user
    ON portfolios(user_id, league_id, season_ref_id)
    WHERE league_id IS NOT NULL;
CREATE INDEX idx_portfolios_league
    ON portfolios(league_id, season_ref_id)
    WHERE league_id IS NOT NULL;
---- create above / drop below ----
DROP INDEX IF EXISTS idx_portfolios_league;
DROP INDEX IF EXISTS idx_portfolios_league_season_user;
DROP INDEX IF EXISTS idx_portfolios_global_user;
ALTER TABLE portfolios DROP COLUMN IF EXISTS season_ref_id;
ALTER TABLE portfolios DROP COLUMN IF EXISTS league_id;
ALTER TABLE portfolios ADD CONSTRAINT portfolios_user_id_season_id_key UNIQUE (user_id, season_id);
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
