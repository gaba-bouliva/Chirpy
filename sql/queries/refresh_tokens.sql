-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTokenByUserId :one
SELECT * FROM refresh_tokens WHERE user_id = $1 LIMIT 1;

-- name: GetToken :one
SELECT * FROM refresh_tokens WHERE token = $1 LIMIT 1;

-- name: RevokeToken :exec
UPDATE refresh_tokens SET revoked_at = $1, updated_at = $2 WHERE token = $3;

-- name: DeleteUserToken :exec
DELETE FROM refresh_tokens WHERE user_id = $1;


-- name: DeleteAllTokens :exec
DELETE FROM refresh_tokens;