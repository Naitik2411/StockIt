-- Write your migrate up statements here
ALTER TABLE portfolios
    ALTER COLUMN season_id TYPE VARCHAR(64);

---- create above / drop below ----

ALTER TABLE portfolios
    ALTER COLUMN season_id TYPE VARCHAR(20);