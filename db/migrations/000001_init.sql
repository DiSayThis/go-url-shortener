-- +goose Up
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT uuidv7(),
    email VARCHAR(254) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    password_hash TEXT,
    role VARCHAR(32) NOT NULL DEFAULT 'user',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_users_public_id UNIQUE (public_id),
    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT chk_users_email_normalized CHECK (
        email = lower(btrim(email))
        AND length(email) > 0
    ),
    CONSTRAINT chk_users_role CHECK (
        role IN ('user', 'moderator', 'admin')
    ),
    CONSTRAINT chk_users_status CHECK (
        status IN ('active', 'blocked', 'pending')
    )
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    family_id UUID NOT NULL DEFAULT uuidv7(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    parent_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,
    created_ip INET,
    last_used_ip INET,
    user_agent TEXT,
    CONSTRAINT fk_refresh_tokens_parent FOREIGN KEY (parent_id) REFERENCES refresh_tokens(id) ON DELETE
    SET
        NULL,
        CONSTRAINT uq_refresh_tokens_hash UNIQUE (token_hash),
        CONSTRAINT uq_refresh_tokens_parent UNIQUE (parent_id),
        CONSTRAINT chk_refresh_tokens_hash_length CHECK (octet_length(token_hash) = 32),
        CONSTRAINT chk_refresh_tokens_expiration CHECK (expires_at > created_at)
);

CREATE INDEX idx_refresh_tokens_user_active ON refresh_tokens (user_id, created_at DESC)
WHERE
    revoked_at IS NULL;

CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family_id);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

CREATE TABLE links (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT uuidv7(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    url TEXT NOT NULL,
    hash VARCHAR(32) NOT NULL,
    title VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_links_public_id UNIQUE (public_id),
    CONSTRAINT uq_links_hash UNIQUE (hash),
    CONSTRAINT chk_links_url_not_empty CHECK (length(btrim(url)) > 0),
    CONSTRAINT chk_links_hash_length CHECK (
        char_length(hash) BETWEEN 6
        AND 32
    ),
    CONSTRAINT chk_links_expiration CHECK (
        expires_at IS NULL
        OR expires_at > created_at
    )
);

CREATE INDEX idx_links_user_active ON links (user_id, created_at DESC)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_links_expires_at ON links (expires_at)
WHERE
    expires_at IS NOT NULL
    AND deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS links;

DROP TABLE IF EXISTS refresh_tokens;

DROP TABLE IF EXISTS users;