CREATE TABLE IF NOT EXISTS oidc_auth_providers (
    oidc_auth_provider_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (name <> ''),
    issuer_domain_path TEXT NOT NULL UNIQUE CHECK (issuer_domain_path <> ''),
    client_id TEXT NOT NULL CHECK (client_id <> ''),
    client_secret TEXT NOT NULL CHECK (client_secret <> '')
);

CREATE TABLE IF NOT EXISTS user_auth_methods (
    user_auth_method_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    oidc_auth_provider_id INTEGER NOT NULL REFERENCES oidc_auth_providers(oidc_auth_provider_id) ON DELETE CASCADE,
    external_subject TEXT NOT NULL CHECK (external_subject <> ''),
    last_oidc_authenticated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (oidc_auth_provider_id, external_subject),
    UNIQUE (user_id, oidc_auth_provider_id)
);

CREATE TABLE IF NOT EXISTS oidc_clients (
    oidc_client_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (name <> ''),
    domain TEXT NOT NULL UNIQUE CHECK (domain <> ''),
    client_id TEXT NOT NULL UNIQUE CHECK (client_id <> ''),
    client_secret TEXT NOT NULL CHECK (client_secret <> '')
);
