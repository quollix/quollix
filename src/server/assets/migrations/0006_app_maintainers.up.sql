CREATE TABLE IF NOT EXISTS app_maintainers (
    name TEXT PRIMARY KEY CHECK (name <> ''),
    public_key_raw BYTEA NOT NULL UNIQUE CHECK (length(public_key_raw) > 0),
    public_key_fingerprint TEXT NOT NULL UNIQUE CHECK (public_key_fingerprint <> ''),
    last_updated_at TIMESTAMPTZ NOT NULL
);
