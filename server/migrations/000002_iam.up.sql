
CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT NOT NULL,
    name           TEXT NOT NULL,
    password_hash  BYTEA NOT NULL,                    
    is_admin       BOOLEAN NOT NULL DEFAULT false,
    is_active      BOOLEAN NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_key ON users (lower(email));

CREATE TABLE sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  BYTEA NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user ON sessions (user_id);

CREATE TYPE app_role AS ENUM ('admin', 'developer');

CREATE TABLE app_members (
    app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        app_role NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (app_id, user_id)
);

CREATE INDEX idx_app_members_user ON app_members (user_id);


ALTER TABLE api_keys
    ADD COLUMN created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE;
