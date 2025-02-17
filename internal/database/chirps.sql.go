// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: chirps.sql

package database

import (
	"context"
	"time"
)

const createChirp = `-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id )
VALUES ($1, $2, $3, $4, $5)
RETURNING id, body, created_at, updated_at, user_id
`

type CreateChirpParams struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Body      string
	UserID    string
}

func (q *Queries) CreateChirp(ctx context.Context, arg CreateChirpParams) (Chirp, error) {
	row := q.db.QueryRowContext(ctx, createChirp,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Body,
		arg.UserID,
	)
	var i Chirp
	err := row.Scan(
		&i.ID,
		&i.Body,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
	)
	return i, err
}

const deleteChirpById = `-- name: DeleteChirpById :exec
DELETE FROM chirps WHERE id = $1
`

func (q *Queries) DeleteChirpById(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteChirpById, id)
	return err
}

const getAllChirps = `-- name: GetAllChirps :many
SELECT id, body, created_at, updated_at, user_id FROM chirps ORDER BY created_at ASC
`

func (q *Queries) GetAllChirps(ctx context.Context) ([]Chirp, error) {
	rows, err := q.db.QueryContext(ctx, getAllChirps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chirp
	for rows.Next() {
		var i Chirp
		if err := rows.Scan(
			&i.ID,
			&i.Body,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllUserChirps = `-- name: GetAllUserChirps :many
SELECT id, body, created_at, updated_at, user_id FROM chirps WHERE user_id = $1 ORDER BY created_at ASC
`

func (q *Queries) GetAllUserChirps(ctx context.Context, userID string) ([]Chirp, error) {
	rows, err := q.db.QueryContext(ctx, getAllUserChirps, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chirp
	for rows.Next() {
		var i Chirp
		if err := rows.Scan(
			&i.ID,
			&i.Body,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getChirpById = `-- name: GetChirpById :one
SELECT id, body, created_at, updated_at, user_id FROM chirps WHERE id = $1 LIMIT 1
`

func (q *Queries) GetChirpById(ctx context.Context, id string) (Chirp, error) {
	row := q.db.QueryRowContext(ctx, getChirpById, id)
	var i Chirp
	err := row.Scan(
		&i.ID,
		&i.Body,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
	)
	return i, err
}
