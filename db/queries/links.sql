-- name: CreateLink :one
INSERT INTO
	links (url, hash)
VALUES
	(sqlc.arg(url), sqlc.arg(hash)) RETURNING id,
	created_at,
	updated_at,
	deleted_at,
	url,
	hash;

-- name: GetLinkByID :one
SELECT
	id,
	created_at,
	updated_at,
	deleted_at,
	url,
	hash
FROM
	links
WHERE
	id = sqlc.arg(id)
	AND deleted_at IS NULL;

-- name: GetLinkByHash :one
SELECT
	id,
	created_at,
	updated_at,
	deleted_at,
	url,
	hash
FROM
	links
WHERE
	hash = sqlc.arg(hash)
	AND deleted_at IS NULL;

-- name: ListLinks :many
SELECT
	id,
	created_at,
	updated_at,
	deleted_at,
	url,
	hash
FROM
	links
WHERE
	deleted_at IS NULL
ORDER BY
	id DESC
LIMIT
	sqlc.arg(page_size) :: integer OFFSET sqlc.arg(page_offset) :: integer;

-- name: UpdateLinkURL :one
UPDATE
	links
SET
	url = sqlc.arg(url),
	updated_at = now()
WHERE
	id = sqlc.arg(id)
	AND deleted_at IS NULL RETURNING id,
	created_at,
	updated_at,
	deleted_at,
	url,
	hash;

-- name: UpdateLinkUrlAndHash :one
UPDATE
	links
SET
	url = sqlc.arg(url),
	hash = sqlc.arg(hash),
	updated_at = now()
WHERE
	id = sqlc.arg(id)
	AND deleted_at IS NULL RETURNING id,
	created_at,
	updated_at,
	deleted_at,
	url,
	hash;

-- name: SoftDeleteLink :execrows
UPDATE
	links
SET
	deleted_at = now(),
	updated_at = now()
WHERE
	id = sqlc.arg(id)
	AND deleted_at IS NULL;

-- name: RestoreLink :execrows
UPDATE
	links
SET
	deleted_at = NULL,
	updated_at = now()
WHERE
	id = sqlc.arg(id)
	AND deleted_at IS NOT NULL;

-- name: HardDeleteLink :execrows
DELETE FROM
	links
WHERE
	id = sqlc.arg(id);