CREATE TABLE apps (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT NOT NULL UNIQUE                 
                CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
    name        TEXT NOT NULL,                       
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,                    -- label, e.g. "ci" or "laptop"
    key_hash    BYTEA NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at  TIMESTAMPTZ
);

-- One row per unique file. Content-addressed: the hash is the identity,
-- and the file lives in R2 at "assets/<hash>".
CREATE TABLE assets (
    hash          TEXT PRIMARY KEY,               -- base64url(sha256), no padding
    key           TEXT NOT NULL,                  -- md5 hex, client cache key (required by expo update protocol)
    content_type  TEXT NOT NULL,
    file_ext      TEXT,                           -- ".png", ".bundle", ...
    size_bytes    BIGINT NOT NULL CHECK (size_bytes >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE platform AS ENUM ('ios', 'android');
CREATE TYPE update_kind AS ENUM ('update', 'rollback_to_embedded');


CREATE TABLE app_platforms (
    app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    platform    platform NOT NULL,
    bundle_id   TEXT,                             
    enabled     BOOLEAN NOT NULL DEFAULT true,    
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (app_id, platform)
);


CREATE TABLE updates (
    id               UUID PRIMARY KEY,            -- manifest "id"
    group_id         UUID NOT NULL,               -- shared by ios + android of one publish
    app_id           UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    channel          TEXT NOT NULL,               -- "production", "staging", ...
    platform         platform NOT NULL,
    runtime_version  TEXT NOT NULL,
    kind             update_kind NOT NULL DEFAULT 'update',
    launch_asset     TEXT REFERENCES assets(hash), -- NULL for rollback_to_embedded
    message          TEXT,
    git_commit       TEXT,
    rollout_percent  SMALLINT NOT NULL DEFAULT 100 CHECK (rollout_percent BETWEEN 0 AND 100),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK ((kind = 'update') = (launch_asset IS NOT NULL)),
    FOREIGN KEY (app_id, platform) REFERENCES app_platforms (app_id, platform)
);

CREATE INDEX idx_updates_lookup
    ON updates (app_id, channel, platform, runtime_version, created_at DESC);

CREATE INDEX idx_updates_group ON updates (group_id);

CREATE TABLE update_assets (
    update_id   UUID NOT NULL REFERENCES updates(id) ON DELETE CASCADE,
    asset_hash  TEXT NOT NULL REFERENCES assets(hash),
    PRIMARY KEY (update_id, asset_hash)
);
