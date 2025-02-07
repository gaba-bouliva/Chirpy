-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1 LIMIT 1;

-- name: DeleteAllChirps :exec
DELETE FROM chirps;

