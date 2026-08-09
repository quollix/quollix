CREATE TABLE IF NOT EXISTS app_secrets (
    app_id INTEGER NOT NULL REFERENCES apps(app_id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name ~ '^SECRET_[A-Z0-9_]+$'),
    value TEXT NOT NULL CHECK (value <> ''),
    PRIMARY KEY (app_id, name)
);
