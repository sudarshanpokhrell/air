DROP INDEX idx_updates_lookup;
CREATE INDEX idx_updates_lookup
    ON updates (app_id, channel, platform, runtime_version, created_at DESC);

ALTER TABLE apps
    DROP COLUMN signing_key_id,
    DROP COLUMN signing_certificate;

ALTER TABLE updates
    DROP COLUMN rolled_back_at,
    DROP COLUMN signature,
    DROP COLUMN manifest;
