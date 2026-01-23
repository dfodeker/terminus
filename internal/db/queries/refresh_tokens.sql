-- db/queries/refresh_tokens.sql

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, token_hash, user_agent, ip_address, expires_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 AND revoked = FALSE AND expires_at > NOW();

-- name: GetRefreshTokenByID :one
SELECT * FROM refresh_tokens WHERE id = $1;

-- name: ListRefreshTokensByUser :many
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND revoked = FALSE
ORDER BY created_at DESC;

-- name: UpdateRefreshTokenLastUsed :exec
UPDATE refresh_tokens
SET last_used_at = NOW()
WHERE id = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked = TRUE, revoked_at = NOW()
WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked = TRUE, revoked_at = NOW()
WHERE user_id = $1 AND revoked = FALSE;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < NOW() OR revoked = TRUE;
