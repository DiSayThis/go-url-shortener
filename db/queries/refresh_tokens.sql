-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id,
    token_hash,
    expires_at,
    created_ip,
    user_agent
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(expires_at),
    sqlc.narg(created_ip),
    sqlc.narg(user_agent)
)
RETURNING *;

-- name: CreateRotatedRefreshToken :one
INSERT INTO refresh_tokens (
    family_id,
    user_id,
    token_hash,
    parent_id,
    expires_at,
    created_ip,
    user_agent
)
VALUES (
    sqlc.arg(family_id),
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(parent_id),
    sqlc.arg(expires_at),
    sqlc.narg(created_ip),
    sqlc.narg(user_agent)
)
RETURNING *;

-- name: GetRefreshTokenForUpdate :one
SELECT *
FROM refresh_tokens
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: MarkRefreshTokenUsed :one
UPDATE refresh_tokens
SET used_at = now(),
    last_used_ip = sqlc.narg(last_used_ip)
WHERE id = sqlc.arg(id)
  AND used_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
RETURNING *;

-- name: RevokeRefreshTokenFamily :execrows
UPDATE refresh_tokens
SET revoked_at = COALESCE(revoked_at, now()),
    revoked_reason = COALESCE(revoked_reason, sqlc.arg(revoked_reason))
WHERE family_id = sqlc.arg(family_id)
  AND revoked_at IS NULL;

-- name: RevokeUserRefreshTokenFamily :execrows
UPDATE refresh_tokens
SET revoked_at = COALESCE(revoked_at, now()),
    revoked_reason = COALESCE(revoked_reason, sqlc.arg(revoked_reason))
WHERE family_id = sqlc.arg(family_id)
  AND user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;

-- name: RevokeAllUserRefreshTokens :execrows
UPDATE refresh_tokens
SET revoked_at = COALESCE(revoked_at, now()),
    revoked_reason = COALESCE(revoked_reason, sqlc.arg(revoked_reason))
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;

-- name: ListActiveUserRefreshSessions :many
SELECT DISTINCT ON (family_id)
    id,
    family_id,
    user_id,
    parent_id,
    created_at,
    expires_at,
    last_used_ip,
    created_ip,
    user_agent
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
  AND expires_at > now()
ORDER BY family_id, created_at DESC;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE FROM refresh_tokens
WHERE expires_at < sqlc.arg(expired_before)
   OR revoked_at < sqlc.arg(revoked_before);
