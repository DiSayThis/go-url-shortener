-- +goose Up

CREATE TABLE links (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    url        VARCHAR(255) NOT NULL,
    hash       VARCHAR(255) NOT NULL
);

CREATE INDEX idx_links_deleted_at
    ON links(deleted_at);

CREATE UNIQUE INDEX idx_links_hash
    ON links(hash);

-- +goose Down

DROP TABLE links;