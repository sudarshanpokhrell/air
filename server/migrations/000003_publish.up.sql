ALTER TABLE updates
    ADD COLUMN manifest       BYTEA,        
    ADD COLUMN signature      TEXT,         
    ADD COLUMN rolled_back_at TIMESTAMPTZ;  

ALTER TABLE updates ALTER COLUMN manifest SET NOT NULL;

ALTER TABLE apps
    ADD COLUMN signing_certificate TEXT,
    ADD COLUMN signing_key_id      TEXT;

DROP INDEX idx_updates_lookup;

CREATE INDEX idx_updates_lookup
    ON updates (app_id, channel, platform, runtime_version, created_at DESC)
    WHERE rolled_back_at IS NULL;
