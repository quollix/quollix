CREATE TABLE IF NOT EXISTS configs (
    key TEXT PRIMARY KEY CHECK (key <> ''),
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE CHECK (username <> ''),
    email TEXT NOT NULL UNIQUE CHECK (email <> ''),
    hashed_password TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL,
    set_password_token TEXT NOT NULL,
    set_password_token_expiration_date TIMESTAMPTZ NOT NULL,
    creation_date TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS users_set_password_token_idx
    ON users (set_password_token)
    WHERE set_password_token <> '';

CREATE TABLE IF NOT EXISTS user_sessions (
    session_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    audience TEXT NOT NULL CHECK (audience <> ''),
    hashed_cookie_value TEXT NOT NULL UNIQUE CHECK (hashed_cookie_value <> ''),
    cookie_expiration_date TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx
    ON user_sessions (user_id);

CREATE INDEX IF NOT EXISTS user_sessions_cookie_expiration_date_idx
    ON user_sessions (cookie_expiration_date);

CREATE TABLE IF NOT EXISTS apps (
    app_id SERIAL PRIMARY KEY,
    maintainer TEXT NOT NULL CHECK (maintainer <> ''),
    app_name TEXT NOT NULL UNIQUE CHECK (app_name <> ''),
    version_name TEXT NOT NULL CHECK (version_name <> ''),
    port TEXT NOT NULL CHECK (port <> ''),
    version_creation_date TIMESTAMPTZ NOT NULL,
    version_content BYTEA NOT NULL UNIQUE CHECK (length(version_content) > 0),
    should_be_running BOOLEAN NOT NULL,
    access_policy TEXT NOT NULL CHECK (access_policy <> ''),
    client_id TEXT NOT NULL UNIQUE CHECK (client_id <> ''),
    client_secret TEXT NOT NULL UNIQUE CHECK (client_secret <> ''),
    automatic_backups_enabled BOOLEAN NOT NULL,
    automatic_updates_enabled BOOLEAN NOT NULL,
    UNIQUE (maintainer, app_name)
);

CREATE TABLE IF NOT EXISTS groups (
    group_id SERIAL PRIMARY KEY,
    group_name TEXT NOT NULL UNIQUE CHECK (group_name <> '')
);

CREATE TABLE IF NOT EXISTS memberships (
    group_id INTEGER NOT NULL REFERENCES groups(group_id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS memberships_user_id_group_id_idx
    ON memberships (user_id, group_id);

CREATE TABLE IF NOT EXISTS app_access (
    group_id INTEGER NOT NULL REFERENCES groups(group_id) ON DELETE CASCADE,
    app_name TEXT NOT NULL CHECK (app_name <> ''),
    PRIMARY KEY (group_id, app_name)
);

CREATE INDEX IF NOT EXISTS app_access_app_name_group_id_idx
    ON app_access (app_name, group_id);
