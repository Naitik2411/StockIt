-- Write your migrate up statements here
CREATE TABLE leagues(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_by UUID NOT NULL references users(id),
    invite_code VARCHAR(8) NOT NULL UNIQUE,
    max_members INT NOT NULL DEFAULT 20,
    is_public boolean NOT NULL DEFAULT false,
    status VARCHAR(10) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_leagues_invite_code on leagues(invite_code);
---- create above / drop below ----

DROP TABLE IF EXISTS leagues;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.    
