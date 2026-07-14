-- Bootstrap migration: enable UUID generation for future schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

---- create above / drop below ----

DROP EXTENSION IF EXISTS "pgcrypto";
