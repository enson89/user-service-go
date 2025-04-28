CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(60)  NOT NULL,
    role            VARCHAR(50)  NOT NULL DEFAULT 'user',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Index on lower(email) to make case-insensitive lookups fast
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));