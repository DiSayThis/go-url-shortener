-- name: CreateLink :one
INSERT INTO links (
    user_id,
    url,
    hash,
    title,
    expires_at
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(url),
    sqlc.arg(hash),
    sqlc.narg(title),
    sqlc.narg(expires_at)
)
RETURNING *;

-- name: GetLinkByID :one
SELECT *
FROM links
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL;

-- name: GetLinkByPublicID :one
SELECT *
FROM links
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL;

-- name: GetLinkByHash :one
SELECT *
FROM links
WHERE hash = sqlc.arg(hash)
  AND deleted_at IS NULL
  AND is_active = true
  AND (expires_at IS NULL OR expires_at > now());

-- name: ListLinksByUser :many
SELECT *
FROM links
WHERE user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size)::integer
OFFSET sqlc.arg(page_offset)::integer;

-- name: UpdateLinkURL :one
UPDATE links
SET url = sqlc.arg(url),
    updated_at = now()
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateLinkUrlAndHash :one
UPDATE links
SET url = sqlc.arg(url),
    hash = sqlc.arg(hash),
    updated_at = now()
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateLinkMetadata :one
UPDATE links
SET title = sqlc.narg(title),
    expires_at = sqlc.narg(expires_at),
    updated_at = now()
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
RETURNING *;

-- name: SetLinkActive :one
UPDATE links
SET is_active = sqlc.arg(is_active),
    updated_at = now()
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteLink :one
UPDATE links
SET deleted_at = now(),
    updated_at = now()
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
RETURNING id;

-- name: RestoreLink :one
UPDATE links
SET deleted_at = NULL,
    updated_at = now()
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
  AND deleted_at IS NOT NULL
RETURNING *;

-- name: HardDeleteLink :one
DELETE FROM links
WHERE public_id = sqlc.arg(public_id)
  AND user_id = sqlc.arg(user_id)
RETURNING id;
