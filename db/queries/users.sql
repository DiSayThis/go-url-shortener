-- name: CreateUser :one
INSERT INTO
  users (email, display_name, password_hash)
VALUES
  (
    sqlc.arg(email),
    sqlc.arg(display_name),
    sqlc.arg(password_hash)
  ) RETURNING *;

-- name: CreateUserWithoutPassword :one
INSERT INTO
  users (
    email,
    display_name,
    email_verified_at
  )
VALUES
  (
    sqlc.arg(email),
    sqlc.arg(display_name),
    sqlc.narg(email_verified_at)
  ) RETURNING *;

-- name: GetUserByID :one
SELECT
  *
FROM
  users
WHERE
  id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: GetUserByPublicID :one
SELECT
  *
FROM
  users
WHERE
  public_id = sqlc.arg(public_id)
  AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = sqlc.arg(email)
  AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE
  users
SET
  display_name = sqlc.arg(display_name),
  updated_at = now()
WHERE
  id = sqlc.arg(id)
  AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserPassword :one
UPDATE
  users
SET
  password_hash = sqlc.arg(password_hash),
  updated_at = now()
WHERE
  id = sqlc.arg(id)
  AND deleted_at IS NULL RETURNING *;

-- name: MarkUserEmailVerified :one
UPDATE
  users
SET
  email_verified_at = COALESCE(email_verified_at, now()),
  updated_at = now()
WHERE
  id = sqlc.arg(id)
  AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserLastLogin :exec
UPDATE
  users
SET
  last_login_at = now(),
  updated_at = now()
WHERE
  id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: UpdateUserRole :one
UPDATE
  users
SET
  role = sqlc.arg(role),
  updated_at = now()
WHERE
  public_id = sqlc.arg(public_id)
  AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserStatus :one
UPDATE
  users
SET
  status = sqlc.arg(status),
  updated_at = now()
WHERE
  public_id = sqlc.arg(public_id)
  AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteUser :one
UPDATE
  users
SET
  status = 'blocked',
  deleted_at = now(),
  updated_at = now()
WHERE
  id = sqlc.arg(id)
  AND deleted_at IS NULL RETURNING id;