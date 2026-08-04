-- Write your migrate up statements here
CREATE TABLE league_members(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(10) NOT NULL DEFAULT 'member'
        CHECK (role in ('admin', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, league_id)
);

CREATE INDEX idx_league_members_user ON league_members(user_id);
CREATE INDEX idx_league_members_league ON league_members(league_id);
---- create above / drop below ----
DROP TABLE IF EXISTS league_members;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
