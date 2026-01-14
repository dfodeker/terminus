-- db/queries/users.sql

-- name: CreateUser :one
INSERT INTO users (
    email, email_verified, password_hash,
    first_name, last_name, avatar_url, phone,
    status, locale, timezone
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    phone = COALESCE(sqlc.narg('phone'), phone),
    locale = COALESCE(sqlc.narg('locale'), locale),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users
SET email = $2, email_verified = FALSE, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: VerifyUserEmail :one
UPDATE users
SET email_verified = TRUE, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserStatus :one
UPDATE users
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL);

-- name: SearchUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND (
    email ILIKE '%' || sqlc.arg('query') || '%'
    OR first_name ILIKE '%' || sqlc.arg('query') || '%'
    OR last_name ILIKE '%' || sqlc.arg('query') || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');
