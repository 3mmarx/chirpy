-- name: CreateRefreshToken :one
INSERT INTO
    refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES ($1, NOW(), NOW(), $2, NULL, $3)
RETURNING *;


-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = Now(), updated_at = Now() WHERE token = $1;
