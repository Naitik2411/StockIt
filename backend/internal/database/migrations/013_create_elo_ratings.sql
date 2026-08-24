-- Write your migrate up statements here
-- One row per user. Updated after every season they participate in.
CREATE TABLE elo_ratings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    rating INT NOT NULL DEFAULT 1000,
    seasons_played INT NOT NULL DEFAULT 0,
    peak_rating INT NOT NULL DEFAULT 1000,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_elo_ratings_rating ON elo_ratings(rating DESC);

---- create above / drop below ----

DROP TABLE IF EXISTS elo_ratings;
